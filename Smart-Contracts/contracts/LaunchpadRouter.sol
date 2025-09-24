// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {LaunchpadFactory} from "./LaunchpadFactory.sol";
import {ILaunchPool} from "./interfaces/ILaunchPool.sol";

/**
 * @title LaunchpadRouter
 * @notice Convenience router for interacting with LaunchPool contracts deployed via LaunchpadFactory.
 */
contract LaunchpadRouter {
    using SafeERC20 for IERC20;

    LaunchpadFactory public immutable factory;
    IERC20 public immutable usdc;

    constructor(address _factory, address _usdc) {
        factory = LaunchpadFactory(_factory);
        usdc = IERC20(_usdc);
    }

    /*//////////////////////////////////////////////////////////////
                                swap helpers
    //////////////////////////////////////////////////////////////*/

    function swapExactUSDCForTokens(address token, uint256 amountIn, uint256 minOut, address to)
        external
        returns (uint256 amountOut)
    {
        address pool = factory.getPool(token);
        require(pool != address(0), "pool not found");
        usdc.safeTransferFrom(msg.sender, pool, amountIn);
        amountOut = ILaunchPool(pool).swapExactUSDCForTokens(amountIn, minOut, to);
    }

    function swapExactTokensForUSDC(address token, uint256 amountIn, uint256 minOut, address to)
        external
        returns (uint256 amountOut)
    {
        address pool = factory.getPool(token);
        require(pool != address(0), "pool not found");
        IERC20(token).safeTransferFrom(msg.sender, pool, amountIn);
        amountOut = ILaunchPool(pool).swapExactTokensForUSDC(amountIn, minOut, to);
    }

    /*//////////////////////////////////////////////////////////////
                             liquidity helpers
    //////////////////////////////////////////////////////////////*/

    function addLiquidity(address token, uint256 amountUSDC, uint256 amountToken, address to)
        external
        returns (uint256 lpMinted)
    {
        address pool = factory.getPool(token);
        require(pool != address(0), "pool not found");
        usdc.safeTransferFrom(msg.sender, pool, amountUSDC);
        IERC20(token).safeTransferFrom(msg.sender, pool, amountToken);
        lpMinted = ILaunchPool(pool).addLiquidity(amountUSDC, amountToken, to);
    }

    function removeLiquidity(address token, uint256 lpTokens, address to)
        external
        returns (uint256 amountUSDC, uint256 amountToken)
    {
        address pool = factory.getPool(token);
        require(pool != address(0), "pool not found");
        // Transfer LP tokens to pool then call remove
        IERC20(pool).safeTransferFrom(msg.sender, pool, lpTokens);
        (amountUSDC, amountToken) = ILaunchPool(pool).removeLiquidity(lpTokens, to);
    }
}
