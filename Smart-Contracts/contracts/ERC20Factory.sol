// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ProjectToken} from "./ProjectToken.sol";

/**
 * @title ERC20Factory
 * @notice Deploys minimal ERC20 tokens with 18 decimals and mints full initial supply to msg.sender.
 */
contract ERC20Factory {
    event TokenCreated(address indexed token, address indexed creator, string name, string symbol, uint256 initialSupply);

    function createToken(string memory name_, string memory symbol_, uint256 initialSupply) external returns (address token) {
        ProjectToken t = new ProjectToken(name_, symbol_, initialSupply, msg.sender);
        token = address(t);
        emit TokenCreated(token, msg.sender, name_, symbol_, initialSupply);
    }
}
