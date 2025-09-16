// contracts/stork/UpgradeableStork.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import "./SelfServeStork.sol";

contract UpgradeableSelfServeStork is Initializable, UUPSUpgradeable, Ownable2StepUpgradeable, SelfServeStork {
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address initialOwner) initializer public {
        __Ownable_init(initialOwner);
        __UUPSUpgradeable_init();
    }

    function createPublisherUser(
        address pubKey,
        uint256 singleUpdateFee
    ) public onlyOwner override {
        addPublisherUser(pubKey, singleUpdateFee);
    }

    function deletePublisherUser(address pubKey) public onlyOwner override {
        removePublisherUser(pubKey);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    function renounceOwnership() public virtual override onlyOwner {
        revert("Ownable: renouncing ownership is disabled");
    }
}
