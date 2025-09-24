// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title ProjectToken
 * @notice Minimal ERC20 (18 decimals) for Paxeer Launchpad projects.
 *         Entire initial supply is minted at deploy time to the provided recipient.
 */
contract ProjectToken is ERC20 {
    constructor(string memory name_, string memory symbol_, uint256 initialSupply, address to) ERC20(name_, symbol_) {
        require(to != address(0), "to=0");
        _mint(to, initialSupply);
    }
}
