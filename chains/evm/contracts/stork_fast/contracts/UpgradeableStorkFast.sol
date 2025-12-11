// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import "./StorkFast.sol";

contract UpgradeableStorkFast is
    Initializable,
    UUPSUpgradeable,
    Ownable2StepUpgradeable,
    StorkFast
{
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address initialOwner,
        address signerAddress,
        uint verificationFeeInWei
    ) public initializer {
        __Ownable_init(initialOwner);
        __UUPSUpgradeable_init();

        _initialize(signerAddress, verificationFeeInWei);
    }

    function updateSignerAddress(
        address signerAddress
    ) public override onlyOwner {
        setSignerAddress(signerAddress);
    }

    function updateVerificationFeeInWei(
        uint verificationFeeInWei
    ) public override onlyOwner {
        setVerificationFeeInWei(verificationFeeInWei);
    }

    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyOwner {}

    function renounceOwnership() public virtual override onlyOwner {
        revert("Ownable: renouncing ownership is disabled");
    }
}
