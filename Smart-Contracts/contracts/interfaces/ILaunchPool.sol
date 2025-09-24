// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {LaunchPoolOracle} from "../LaunchPoolOracle.sol";

/**
 * @title ILaunchPool
 * @dev Minimal interface for interacting with Paxeer LaunchPad pools.
 */
interface ILaunchPool {
    /*//////////////////////////////////////////////////////////////
                                 view functions
    //////////////////////////////////////////////////////////////*/
    function token() external view returns (IERC20);

    function usdc() external view returns (IERC20);

    function virtualReserveUSDC() external view returns (uint256);

    function virtualReserveToken() external view returns (uint256);

    function FLOOR_PRICE_X18() external view returns (uint256);

    function oracle() external view returns (LaunchPoolOracle);

    function creator() external view returns (address);

    function treasury() external view returns (address);

    /**
     * @notice Returns current real reserves (excluding virtual reserves)
     */
    function getRealReserves() external view returns (uint256 realUSDC, uint256 realToken);

    /**
     * @notice Returns current spot price (USDC per token scaled by 1e18)
     */
    function currentPriceX18() external view returns (uint256);

    /**
     * @notice Returns a snapshot of the pool state for frontends
     */
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
        );

    /**
     * @notice Calculates the amount of `token` that would be received for a given USDC input.
     */
    function quoteUSDCToToken(uint256 amountIn) external view returns (uint256 amountOut);

    /**
     * @notice Calculates the amount of USDC that would be received for a given `token` input.
     */
    function quoteTokenToUSDC(uint256 amountIn) external view returns (uint256 amountOut);

    /*//////////////////////////////////////////////////////////////
                                 swapping
    //////////////////////////////////////////////////////////////*/

    function swapExactUSDCForTokens(
        uint256 amountIn,
        uint256 minOut,
        address to
    ) external returns (uint256 amountOut);

    function swapExactTokensForUSDC(
        uint256 amountIn,
        uint256 minOut,
        address to
    ) external returns (uint256 amountOut);

    /*//////////////////////////////////////////////////////////////
                                liquidity
    //////////////////////////////////////////////////////////////*/

    function addLiquidity(
        uint256 amountUSDC,
        uint256 amountToken,
        address to
    ) external returns (uint256 lpTokensMinted);

    function removeLiquidity(
        uint256 lpTokens,
        address to
    ) external returns (uint256 amountUSDC, uint256 amountToken);
}
