// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract LoanDistribution {
    struct Loan {
        bytes32 loanID;
        address borrower;
        uint256 principal;
        uint256 interestRate; // Scaled by 1e4 (e.g., 12% = 1200)
        uint256 tenure; // Months
        uint256 emi;
        uint256 paidEMIs;
        uint256 lastPayment;
        bool active;
    }

    mapping(bytes32 => Loan) public loans;
    address public keeper; // Chainlink Keeper address

    event LoanCreated(bytes32 indexed loanID);
    event EMIPaid(bytes32 indexed loanID);

    constructor(address _keeper) {
        keeper = _keeper;
    }

    // Math: EMI = [P × R × (1 + R)^N] / [(1 + R)^N - 1]
    // P = principal, R = monthly rate (annual / 12 / 10000), N = tenure
    function calculateEMI(uint256 principal, uint256 rate, uint256 tenure) internal pure returns (uint256) {
        uint256 r = rate * 1e14 / 12 / 10000; // Monthly rate scaled
        uint256 n = tenure;
        uint256 numerator = principal * r * (1e18 + r) ** n;
        uint256 denominator = ((1e18 + r) ** n) - 1e18;
        return numerator / denominator;
    }

    function createLoan(bytes32 loanID, uint256 principal, uint256 interestRate, uint256 tenure) external payable {
        uint256 emi = calculateEMI(principal, interestRate, tenure);
        loans[loanID] = Loan(loanID, msg.sender, principal, interestRate, tenure, emi, 0, block.timestamp, true);
        emit LoanCreated(loanID);
    }

    function payEMI(bytes32 loanID) external payable {
        Loan storage loan = loans[loanID];
        require(loan.active, "Inactive");
        require(msg.value >= loan.emi, "Insufficient");

        loan.paidEMIs++;
        loan.lastPayment = block.timestamp;
        if (loan.paidEMIs == loan.tenure) loan.active = false;
        emit EMIPaid(loanID);
    }

    // Automation: Chainlink Keeper checks overdue EMIs
    function checkUpkeep(bytes32 loanID) external view returns (bool) {
        Loan storage loan = loans[loanID];
        return loan.active && (block.timestamp - loan.lastPayment) > 30 days;
    }

    function performUpkeep(bytes32 loanID) external {
        require(msg.sender == keeper, "Only keeper");
        Loan storage loan = loans[loanID];
        if ((block.timestamp - loan.lastPayment) > 30 days) {
            loan.active = false; // Default on overdue
        }
    }
}
