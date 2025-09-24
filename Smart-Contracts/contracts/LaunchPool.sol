// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {ILaunchPool} from "./interfaces/ILaunchPool.sol";
import {LaunchPoolOracle} from "./LaunchPoolOracle.sol";

/**
 * @title LaunchPool
 * @notice Constant-product AMM with fixed virtual reserves providing a floor price.
 *         External liquidity may be added post-launch via LP tokens.
 */
contract LaunchPool is ERC20, ReentrancyGuard, ILaunchPool {
    using SafeERC20 for IERC20;

    /*//////////////////////////////////////////////////////////////
                                immutables
    //////////////////////////////////////////////////////////////*/

    IERC20 public immutable override token;
    IERC20 public immutable override usdc;  // USDC (6 decimals assumed)

    uint256 public immutable override virtualReserveUSDC;   // 10_000 USDC * 1e6
    uint256 public immutable override virtualReserveToken;  // 1_000_000_000 tokens * 1e18

    uint256 public FLOOR_PRICE_X18; // baseline/floor price scaled 1e18 (USDC per token)

    address public immutable override creator;
    address public immutable override treasury;

    LaunchPoolOracle public immutable oracle;

    
    /*//////////////////////////////////////////////////////////////
                                  storage
    //////////////////////////////////////////////////////////////*/

    uint256 private reserveUSDC;  // real USDC reserve (excluding virtual)
    uint256 private reserveToken; // real token reserve (excluding virtual)

    uint256 public pendingCreatorFeesUSDC; // fees awaiting collection by creator

    bool private _seeded; // one-time flag for initial token seeding

    /*//////////////////////////////////////////////////////////////
                                    events
    //////////////////////////////////////////////////////////////*/

    event Swap(address indexed sender, uint256 amountIn, uint256 amountOut, bool usdcToToken, address indexed to);
    event AddLiquidity(address indexed provider, uint256 amountUSDC, uint256 amountToken, uint256 lpMinted);
    event RemoveLiquidity(address indexed provider, uint256 lpBurned, uint256 amountUSDC, uint256 amountToken);
    event CollectCreatorFees(uint256 amountUSDC);
    event InitialTokenSeeded(uint256 amount);
    event Sync(uint256 reserveUSDC, uint256 reserveToken);
    event PriceUpdate(uint256 priceX18, uint256 floorX18);

    /*//////////////////////////////////////////////////////////////
                                  constructor
    //////////////////////////////////////////////////////////////*/

    constructor(
        address _token,
        address _usdc,
        address _creator,
        address _treasury,
        uint256 _virtualUSDC,
        uint256 _virtualToken,
        string memory _name,
        string memory _symbol
    ) ERC20(_name, _symbol) {
        require(_token != _usdc, "Identical addresses");
        require(_creator != address(0) && _treasury != address(0), "Zero address");

        token = IERC20(_token);
        usdc = IERC20(_usdc);
        creator = _creator;
        treasury = _treasury;

        virtualReserveUSDC = _virtualUSDC;
        virtualReserveToken = _virtualToken; // pass 0 to enforce single-sided virtual liquidity

        // Floor will be set on seedInitialToken based on (virtualReserveUSDC / seededToken)
        FLOOR_PRICE_X18 = 0;

        oracle = new LaunchPoolOracle(FLOOR_PRICE_X18);
    }

    /// @notice One-time initialization to record the initial token reserve transferred by the factory
    /// @dev Must be called before any swaps or liquidity operations. Idempotent via _seeded flag.
    function seedInitialToken(uint256 amount) external nonReentrant {
        require(!_seeded, "seeded");
        require(reserveToken == 0 && reserveUSDC == 0, "started");
        require(amount > 0, "amount 0");
        // Ensure the contract actually holds at least this much
        uint256 bal = token.balanceOf(address(this));
        require(bal >= amount, "insufficient seed bal");
        reserveToken = amount;
        _seeded = true;
        // Set floor as virtualUSDC / seededToken
        FLOOR_PRICE_X18 = (virtualReserveUSDC * 1e18) / amount;
        uint256 priceAfter = _currentPriceX18(reserveUSDC, reserveToken);
        oracle.update(priceAfter);
        emit PriceUpdate(priceAfter, FLOOR_PRICE_X18);
        emit Sync(reserveUSDC, reserveToken);
        emit InitialTokenSeeded(amount);
    }

    /*//////////////////////////////////////////////////////////////
                                  view helpers
    //////////////////////////////////////////////////////////////*/

    function getRealReserves() public view override returns (uint256 realUSDC, uint256 realToken) {
        realUSDC = reserveUSDC;
        realToken = reserveToken;
    }

    function _currentPriceX18(uint256 _reserveUSDC, uint256 _reserveToken) internal view returns (uint256) {
        return ((virtualReserveUSDC + _reserveUSDC) * 1e18) / (virtualReserveToken + _reserveToken);
    }

    /// @notice Returns current spot price (USDC per token scaled by 1e18)
    function currentPriceX18() public view returns (uint256) {
        return _currentPriceX18(reserveUSDC, reserveToken);
    }

    /// @notice Returns a snapshot of the pool state for frontends
    function getState()
        external
        view
        returns (
            uint256 vUSDC,
            uint256 rUSDC,
            uint256 rToken,
            uint256 spotX18,
            uint256 floorX18,
            uint256 pendingFeesUSDC
        )
    {
        vUSDC = virtualReserveUSDC;
        rUSDC = reserveUSDC;
        rToken = reserveToken;
        spotX18 = _currentPriceX18(reserveUSDC, reserveToken);
        floorX18 = FLOOR_PRICE_X18;
        pendingFeesUSDC = pendingCreatorFeesUSDC;
    }

    /*//////////////////////////////////////////////////////////////
                                    swapping
    //////////////////////////////////////////////////////////////*/

    uint256 private constant FEE_BPS = 100; // 1% = 100 basis points out of 10_000
    uint256 private constant BPS_DENOM = 10_000;

    function swapExactUSDCForTokens(uint256 amountIn, uint256 minOut, address to)
        external
        override
        nonReentrant
        returns (uint256 amountOut)
    {
        require(amountIn > 0, "amountIn 0");
        require(to != address(0), "to 0");

        // Detect actual USDC in (supports router pre-transfer)
        uint256 accountedUSDC = reserveUSDC + pendingCreatorFeesUSDC;
        uint256 balanceNow = usdc.balanceOf(address(this));
        uint256 detectedIn = balanceNow - accountedUSDC;
        require(detectedIn > 0, "No USDC in");

        // Apply 1% fee
        uint256 fee = (detectedIn * FEE_BPS) / BPS_DENOM; // 1% fee
        uint256 amountInAfterFee = detectedIn - fee;

        // Update fee accounting
        uint256 creatorShare = (fee * 75) / 100; // 75%
        uint256 treasuryShare = fee - creatorShare; // 25%
        pendingCreatorFeesUSDC += creatorShare;
        usdc.safeTransfer(treasury, treasuryShare);

        // constant product formula to compute amountOut
        uint256 newReserveUSDC = reserveUSDC + amountInAfterFee;
        uint256 k = (virtualReserveUSDC + reserveUSDC) * (virtualReserveToken + reserveToken);
        uint256 newReserveTokenPlusVirtual = k / (virtualReserveUSDC + newReserveUSDC);
        uint256 newReserveToken = newReserveTokenPlusVirtual - virtualReserveToken;

        // amount out = reserveToken - newReserveToken
        amountOut = reserveToken - newReserveToken;
        require(amountOut >= minOut, "Slippage");

        // update reserves
        reserveUSDC = newReserveUSDC;
        reserveToken = newReserveToken;

        // Compute price and update oracle (no floor enforcement on buy path)
        uint256 priceAfter = _currentPriceX18(reserveUSDC, reserveToken);

        // Transfer tokens to recipient
        token.safeTransfer(to, amountOut);

        // Oracle & events
        oracle.update(priceAfter);
        emit PriceUpdate(priceAfter, FLOOR_PRICE_X18);
        emit Sync(reserveUSDC, reserveToken);
        emit Swap(msg.sender, detectedIn, amountOut, true, to);
    }

    // ---------------------------------------------------------
    // QUOTE helpers
    // ---------------------------------------------------------

    function quoteUSDCToToken(uint256 amountIn) external view override returns (uint256 amountOut) {
        // amountOut calculation without fees
        uint256 newReserveUSDC = reserveUSDC + amountIn;
        uint256 k = (virtualReserveUSDC + reserveUSDC) * (virtualReserveToken + reserveToken);
        uint256 newReserveTokenPlusVirtual = k / (virtualReserveUSDC + newReserveUSDC);
        uint256 newReserveToken = newReserveTokenPlusVirtual - virtualReserveToken;
        amountOut = reserveToken - newReserveToken;
    }

    function quoteTokenToUSDC(uint256 amountIn) external view override returns (uint256 amountOut) {
        // If there is no real USDC, a sell cannot withdraw USDC
        if (reserveUSDC == 0) {
            return 0;
        }

        uint256 newReserveToken = reserveToken + amountIn;
        uint256 k = (virtualReserveUSDC + reserveUSDC) * (virtualReserveToken + reserveToken);
        uint256 denom = virtualReserveToken + newReserveToken;
        uint256 newReserveUSDCPlusVirtual = k / denom;

        // Saturate at zero if virtual portion would exceed new total
        if (newReserveUSDCPlusVirtual <= virtualReserveUSDC) {
            return 0;
        }

        uint256 newReserveUSDC = newReserveUSDCPlusVirtual - virtualReserveUSDC;
        if (reserveUSDC <= newReserveUSDC) {
            return 0;
        }
        amountOut = reserveUSDC - newReserveUSDC;
    }

    // ---------------------------------------------------------
    // SWAP FUNCTIONS
    // ---------------------------------------------------------

    function swapExactTokensForUSDC(uint256 amountIn, uint256 minOut, address to)
        external
        override
        nonReentrant
        returns (uint256 amountOut)
    {
        require(amountIn > 0, "amountIn 0");

        // Detect actual token in (supports router pre-transfer)
        uint256 balanceBeforeT = reserveToken;
        uint256 balanceNowT = token.balanceOf(address(this));
        uint256 detectedInT = balanceNowT - balanceBeforeT;
        require(detectedInT > 0, "No token in");

        uint256 k = (virtualReserveUSDC + reserveUSDC) * (virtualReserveToken + reserveToken);
        uint256 newReserveToken = reserveToken + detectedInT;
        uint256 newReserveUSDCPlusVirtual = k / (virtualReserveToken + newReserveToken);
        require(newReserveUSDCPlusVirtual > virtualReserveUSDC, "No USDC liquidity");
        uint256 newReserveUSDC = newReserveUSDCPlusVirtual - virtualReserveUSDC;

        amountOut = reserveUSDC - newReserveUSDC;
        require(amountOut > 0, "Insufficient out");

        // Apply 1% fee on output USDC
        uint256 feeT = (amountOut * FEE_BPS) / BPS_DENOM;
        uint256 creatorShare = (feeT * 75) / 100;
        uint256 treasuryShare = feeT - creatorShare;
        uint256 amountOutAfterFee = amountOut - feeT;
        require(amountOutAfterFee >= minOut, "Slippage");

        // Update reserves
        reserveToken = newReserveToken;
        reserveUSDC = newReserveUSDC;

        pendingCreatorFeesUSDC += creatorShare;
        usdc.safeTransfer(treasury, treasuryShare);
        usdc.safeTransfer(to, amountOutAfterFee);

        uint256 priceAfter = _currentPriceX18(reserveUSDC, reserveToken);
        require(priceAfter >= FLOOR_PRICE_X18, "Below floor");

        oracle.update(priceAfter);
        emit PriceUpdate(priceAfter, FLOOR_PRICE_X18);
        emit Sync(reserveUSDC, reserveToken);
        emit Swap(msg.sender, detectedInT, amountOutAfterFee, false, to);
    }

    /*//////////////////////////////////////////////////////////////
                               liquidity provision
    //////////////////////////////////////////////////////////////*/

    function addLiquidity(uint256 amountUSDC, uint256 amountToken, address to)
        external
        override
        nonReentrant
        returns (uint256 lpTokensMinted)
    {
        // Detect deltas (supports router pre-transfer)
        uint256 usdcIn = usdc.balanceOf(address(this)) - reserveUSDC;
        uint256 tokenIn = token.balanceOf(address(this)) - reserveToken;
        require(usdcIn > 0 && tokenIn > 0, "amounts 0");

        uint256 _totalSupply = totalSupply();
        if (_totalSupply == 0) {
            // first LP gets shares equivalent to sqrt(real reserves contribution)
            lpTokensMinted = _sqrt(usdcIn * tokenIn);
        } else {
            lpTokensMinted = min(
                (usdcIn * _totalSupply) / reserveUSDC,
                (tokenIn * _totalSupply) / reserveToken
            );
        }
        require(lpTokensMinted > 0, "LP zero");
        _mint(to, lpTokensMinted);

        reserveUSDC += usdcIn;
        reserveToken += tokenIn;

        uint256 priceAfter = _currentPriceX18(reserveUSDC, reserveToken);
        oracle.update(priceAfter);
        emit PriceUpdate(priceAfter, FLOOR_PRICE_X18);
        emit Sync(reserveUSDC, reserveToken);

        emit AddLiquidity(msg.sender, amountUSDC, amountToken, lpTokensMinted);
    }

    function removeLiquidity(uint256 lpTokens, address to)
        external
        override
        nonReentrant
        returns (uint256 amountUSDC, uint256 amountToken)
    {
        require(lpTokens > 0, "LP 0");
        uint256 _totalSupply = totalSupply();
        _burn(msg.sender, lpTokens);

        amountUSDC = (lpTokens * reserveUSDC) / _totalSupply;
        amountToken = (lpTokens * reserveToken) / _totalSupply;

        require(amountUSDC > 0 && amountToken > 0, "Amounts 0");

        reserveUSDC -= amountUSDC;
        reserveToken -= amountToken;

        usdc.safeTransfer(to, amountUSDC);
        token.safeTransfer(to, amountToken);

        uint256 priceAfter = _currentPriceX18(reserveUSDC, reserveToken);
        oracle.update(priceAfter);
        emit PriceUpdate(priceAfter, FLOOR_PRICE_X18);
        emit Sync(reserveUSDC, reserveToken);

        emit RemoveLiquidity(msg.sender, lpTokens, amountUSDC, amountToken);
    }

    /*//////////////////////////////////////////////////////////////
                               creator fee withdrawal
    //////////////////////////////////////////////////////////////*/

    function collectCreatorFees() external nonReentrant {
        require(msg.sender == creator, "Not creator");
        uint256 amount = pendingCreatorFeesUSDC;
        require(amount > 0, "Nothing to collect");
        pendingCreatorFeesUSDC = 0;
        usdc.safeTransfer(creator, amount);
        emit CollectCreatorFees(amount);
    }

    /*//////////////////////////////////////////////////////////////
                               internal helpers
    //////////////////////////////////////////////////////////////*/
    function min(uint256 a, uint256 b) private pure returns (uint256) {
        return a < b ? a : b;
    }

    function _sqrt(uint256 y) private pure returns (uint256 z) {
        if (y > 3) {
            z = y;
            uint256 x = y / 2 + 1;
            while (x < z) {
                z = x;
                x = (y / x + x) / 2;
            }
        } else if (y != 0) {
            z = 1;
        }
    }
}
