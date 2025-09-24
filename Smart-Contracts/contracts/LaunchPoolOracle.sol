// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title LaunchPoolOracle
 * @dev Lightweight time-weighted average price oracle for a LaunchPool.
 * Stores cumulative price and allows off-chain indexers to read updates via events.
 */
contract LaunchPoolOracle {
    struct Observation {
        uint32 timestamp;
        uint256 priceCumulative; // sum(price * time) scaled by 1e18 (USDC per token)
    }

    Observation public lastObservation;

    event OracleUpdate(uint256 priceCumulative, uint32 timestamp);

    constructor(uint256 _initialPrice) {
        // initialize observation with current block timestamp
        lastObservation = Observation({timestamp: uint32(block.timestamp), priceCumulative: _initialPrice});
        emit OracleUpdate(_initialPrice, uint32(block.timestamp));
    }

    /**
     * @dev Called by LaunchPool on every state-changing action to push the running price cumulative.
     * @param currentPrice Current spot price (USDC per token scaled by 1e18)
     */
    function update(uint256 currentPrice) external {
        Observation memory obs = lastObservation;
        uint32 nowTs = uint32(block.timestamp);
        uint32 timeElapsed = nowTs - obs.timestamp;
        // accumulate price * time elapsed
        uint256 newCumulative = obs.priceCumulative + currentPrice * timeElapsed;
        lastObservation = Observation({timestamp: nowTs, priceCumulative: newCumulative});
        emit OracleUpdate(newCumulative, nowTs);
    }
}
