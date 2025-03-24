// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract CrossBorderPayments is ReentrancyGuard {
    struct Payment {
        bytes32 paymentID;
        address sender;
        address receiver;
        uint256 amount;
        bytes32 currencyPair;
        mapping(address => uint256) oracleRates; // Oracle -> Rate (scaled 1e18)
        uint256 aggregatedRate;
        bool settled;
        uint64 timestamp;
    }

    mapping(bytes32 => Payment) public payments;
    address[] public oracles;

    event PaymentInitiated(bytes32 indexed paymentID, address sender, address receiver);
    event PaymentSettled(bytes32 indexed paymentID, uint256 settledAmount);

    constructor(address[] memory _oracles) {
        oracles = _oracles;
    }

    function initiatePayment(bytes32 paymentID, address receiver, bytes32 currencyPair) external payable {
        require(msg.value > 0, "Amount required");
        Payment storage p = payments[paymentID];
        p.sender = msg.sender;
        p.receiver = receiver;
        p.amount = msg.value;
        p.currencyPair = currencyPair;
        p.timestamp = uint64(block.timestamp);
        emit PaymentInitiated(paymentID, msg.sender, receiver);
    }

    // Automation: Called by Chainlink oracles
    function submitForexRate(bytes32 paymentID, uint256 rate) external nonReentrant {
        bool isOracle = false;
        for (uint i = 0; i < oracles.length; i++) {
            if (msg.sender == oracles[i]) isOracle = true;
        }
        require(isOracle, "Not an oracle");

        Payment storage p = payments[paymentID];
        require(!p.settled, "Settled");
        p.oracleRates[msg.sender] = rate;

        if (countOracles(paymentID) >= 3) {
            p.aggregatedRate = aggregateRates(paymentID);
            settlePayment(paymentID);
        }
    }

    // Math: Weighted average of rates (assuming equal weights for simplicity)
    // AggregatedRate = (Σ Rate_i) / n, where n is the number of oracles
    function aggregateRates(bytes32 paymentID) internal view returns (uint256) {
        Payment storage p = payments[paymentID];
        uint256 totalRate = 0;
        uint8 count = 0;
        for (uint i = 0; i < oracles.length; i++) {
            if (p.oracleRates[oracles[i]] > 0) {
                totalRate += p.oracleRates[oracles[i]];
                count++;
            }
        }
        return count > 0 ? totalRate / count : 0;
    }

    function countOracles(bytes32 paymentID) internal view returns (uint8) {
        Payment storage p = payments[paymentID];
        uint8 count = 0;
        for (uint i = 0; i < oracles.length; i++) {
            if (p.oracleRates[oracles[i]] > 0) count++;
        }
        return count;
    }

    function settlePayment(bytes32 paymentID) internal {
        Payment storage p = payments[paymentID];
        uint256 settledAmount = (p.amount * p.aggregatedRate) / 1e18;
        p.settled = true;
        (bool success, ) = p.receiver.call{value: settledAmount}("");
        require(success, "Transfer failed");
        emit PaymentSettled(paymentID, settledAmount);
    }
}
