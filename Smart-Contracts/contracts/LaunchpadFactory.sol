// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {LaunchPool} from "./LaunchPool.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title LaunchpadFactory
 * @notice Deploys LaunchPool contracts for project tokens.
 */
contract LaunchpadFactory is Ownable {
    /*//////////////////////////////////////////////////////////////
                                  state
    //////////////////////////////////////////////////////////////*/

    IERC20 public immutable usdc;
    address public immutable treasury; // protocol treasury

    mapping(address => address) public getPool; // token -> pool
    address[] public allPools;

    event PoolCreated(address indexed token, address pool, address oracle);

    constructor(address _usdc, address _treasury) Ownable(msg.sender) {
        require(_usdc != address(0) && _treasury != address(0), "zero");
        usdc = IERC20(_usdc);
        treasury = _treasury;
    }

    /**
     * @notice Creates a LaunchPool for `_token`. Creator must have approved 1B tokens for the pool to pull.
     *         Virtual liquidity is single-sided: only virtual USDC (10,000 * 1e6). No virtual tokens.
     */
    function createPool(address _token) external returns (address pool) {
        require(getPool[_token] == address(0), "exists");

        // constants per specification
        uint256 virtualUSDC = 10_000 * 1e18; // Paxeer USDC uses 18 decimals
        uint256 seedTokens = 1_000_000_000 * 1e18; // 1B tokens (assumes 18 decimals)

        // ensure creator approved tokens for seeding the pool's real token reserve
        require(IERC20(_token).allowance(msg.sender, address(this)) >= seedTokens, "token allowance");

        string memory name = string.concat("Launch-LP-", _toString(_token));
        string memory symbol = "xLP";

        // Pass 0 as virtualToken to enforce single-sided virtual liquidity (USDC only)
        pool = address(new LaunchPool(_token, address(usdc), msg.sender, treasury, virtualUSDC, 0, name, symbol));

        // pull project tokens from creator into the pool
        IERC20(_token).transferFrom(msg.sender, pool, seedTokens);

        // initialize pool's real token reserve accounting
        LaunchPool(pool).seedInitialToken(seedTokens);

        getPool[_token] = pool;
        allPools.push(pool);

        emit PoolCreated(_token, pool, address(LaunchPool(pool).oracle()));
    }

    function allPoolsLength() external view returns (uint256) {
        return allPools.length;
    }

    /*//////////////////////////////////////////////////////////////
                                 helpers
    //////////////////////////////////////////////////////////////*/
    function _toString(address account) internal pure returns (string memory) {
        return _toHex(abi.encodePacked(account));
    }

    function _toHex(bytes memory data) internal pure returns (string memory) {
        bytes16 hexAlphabet = "0123456789abcdef";
        bytes memory str = new bytes(2 + data.length * 2);
        str[0] = "0";
        str[1] = "x";
        for (uint256 i = 0; i < data.length; i++) {
            str[2 + i * 2] = hexAlphabet[uint8(data[i] >> 4)];
            str[3 + i * 2] = hexAlphabet[uint8(data[i] & 0x0f)];
        }
        return string(str);
    }
}
