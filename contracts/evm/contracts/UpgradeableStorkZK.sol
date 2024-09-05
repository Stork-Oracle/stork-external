// SPDX-License-Identifier: MIT
pragma solidity >=0.8.24 <0.9.0;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "./Stork.sol";

contract UpgradeableStorkZK is Initializable, UUPSUpgradeable, OwnableUpgradeable, Stork {
    function initialize(address storkPublicKey, uint validTimePeriodSeconds, uint singleUpdateFeeInWei) initializer public {
        __Ownable_init();
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
}
