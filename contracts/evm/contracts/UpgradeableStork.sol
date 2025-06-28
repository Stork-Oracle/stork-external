// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import "./Stork.sol";

contract UpgradeableStork is Initializable, UUPSUpgradeable, Ownable2StepUpgradeable, Stork {
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address initialOwner, address storkPublicKey, uint validTimePeriodSeconds, uint singleUpdateFeeInWei) initializer public {
        __Ownable_init(initialOwner);
        __UUPSUpgradeable_init();

        _initialize(storkPublicKey, validTimePeriodSeconds, singleUpdateFeeInWei);
    }

    function updateValidTimePeriodSeconds(uint validTimePeriodSeconds) public onlyOwner override {
        setValidTimePeriodSeconds(validTimePeriodSeconds);
    }

    function updateSingleUpdateFeeInWei(uint maxStorkPerBlock) public onlyOwner override {
        setSingleUpdateFeeInWei(maxStorkPerBlock);
    }

    function updateStorkPublicKey(address storkPublicKey) public onlyOwner override {
        setStorkPublicKey(storkPublicKey);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    function renounceOwnership() public virtual override onlyOwner {
        revert("Ownable: renouncing ownership is disabled");
    }
}
