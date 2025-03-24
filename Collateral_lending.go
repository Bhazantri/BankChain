// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract CollateralizedLending {
    struct Loan {
        bytes32 loanID;
        address borrower;
        uint256 loanAmount;
        uint256 collateralAmount;
        uint256 threshold; // LTV threshold (e.g., 80% = 8e17)
        uint256 currentValue;
        bool active;
    }

    mapping(bytes32 => Loan) public loans;
    address public oracle;

    event LoanCreated(bytes32 indexed loanID);
    event MarginCall(bytes32 indexed loanID);

    constructor(address _oracle) {
        oracle = _oracle;
    }

    function createLoan(bytes32 loanID, uint256 loanAmount, uint256 threshold) external payable {
        loans[loanID] = Loan(loanID, msg.sender, loanAmount, msg.value, threshold, msg.value, true);
        emit LoanCreated(loanID);
    }

    // Automation: Oracle updates collateral value
    // Math: LTV = (Loan Amount / Collateral Value) * 100
    function updateCollateralValue(bytes32 loanID, uint256 newValue) external {
        require(msg.sender == oracle, "Only oracle");
        Loan storage loan = loans[loanID];
        require(loan.active, "Inactive");

        loan.currentValue = newValue;
        uint256 ltv = (loan.loanAmount * 1e18) / newValue;
        if (ltv > loan.threshold) {
            loan.active = false;
            emit MarginCall(loanID);
        }
    }
}
