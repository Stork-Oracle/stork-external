use fuels::{
    prelude::*,
    types::{
        Bits256,
        EvmAddress,
    },
    programs::responses::CallResponse,
};

// Load abi from json
abigen!(Contract(
    name = "Stork",
    abi = "out/debug/stork-abi.json"
));

// sets up the test environment, returns two wallets (first is the stork owner, second is not), the stork contract instance and its id
async fn setup_tests() -> (WalletUnlocked, WalletUnlocked, Stork<WalletUnlocked>) {
    let mut wallets = launch_custom_provider_and_get_wallets(
        WalletsConfig::new(
            Some(2),             /* two wallets */
            Some(1),             /* one coin per wallet */
            Some(1_000), /* 1000 amount per coin */
        ),
        None,
        None,
    )
    .await
    .unwrap();
    let stork_owner = wallets.pop().unwrap();
    let not_stork_owner = wallets.pop().unwrap();

    let stork_instance = get_contract_instance(stork_owner.clone()).await;

    (stork_owner, not_stork_owner, stork_instance)
}

async fn get_contract_instance(wallet: WalletUnlocked) -> Stork<WalletUnlocked> {
    // Launch a local network and deploy the contract
    
    let id = Contract::load_from(
        "./out/debug/stork.bin",
        LoadConfiguration::default(),
    )
    .unwrap()
    .deploy(&wallet, TxPolicies::default())
    .await
    .unwrap();

    let instance = Stork::new(id.clone(), wallet);

    instance
}

async fn initialize_stork_default(stork_instance: &Stork<WalletUnlocked>, stork_owner: WalletUnlocked) -> Result<CallResponse<()>> {
    // construct evm address
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);
    let fee = 1; 
    stork_instance.methods().initialize(
        stork_owner.address().into(),
        stork_pub_key,
        fee,
    ).call().await
}

#[tokio::test]
async fn test_initialization() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    // construct evm address
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);
    let fee = 1;

    stork_instance.methods().initialize(
        stork_owner.address().into(),
        stork_pub_key,
        fee,
    ).call().await.unwrap();

    assert_eq!(stork_instance.methods().stork_public_key().call().await.unwrap().value, stork_pub_key);
    assert_eq!(stork_instance.methods().single_update_fee_in_wei().call().await.unwrap().value, fee);
    assert_eq!(stork_instance.methods().owner().call().await.unwrap().value, State::Initialized(stork_owner.address().into()));
}

#[tokio::test]
async fn test_no_duplicate_initialization() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    // construct evm address
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);
    let fee = 1;

    stork_instance.methods().initialize(
        stork_owner.address().into(),
        stork_pub_key,
        fee,
    ).call().await.unwrap();

    assert_eq!(stork_instance.methods().stork_public_key().call().await.unwrap().value, stork_pub_key);
    assert_eq!(stork_instance.methods().single_update_fee_in_wei().call().await.unwrap().value, fee);
    assert_eq!(stork_instance.methods().owner().call().await.unwrap().value, State::Initialized(stork_owner.address().into()));

    let res = stork_instance.methods().initialize(
        stork_owner.address().into(),
        stork_pub_key,
        fee,
    ).call().await;
    assert!(res.is_err());
}

#[tokio::test]
async fn test_update_stork_public_key_owner() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // construct new public key
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d45").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);

    stork_instance.methods().update_stork_public_key(stork_pub_key).call().await.unwrap();

    assert_eq!(stork_instance.methods().stork_public_key().call().await.unwrap().value, stork_pub_key);
}

#[tokio::test]
async fn test_update_stork_public_key_not_owner() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // get current public key
    let current_stork_pub_key = stork_instance.methods().stork_public_key().call().await.unwrap().value;

    // construct new public key
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d45").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);

    let res = stork_instance.clone().with_account(not_stork_owner).methods().update_stork_public_key(stork_pub_key).call().await;
    assert!(res.is_err());
    assert_eq!(stork_instance.methods().stork_public_key().call().await.unwrap().value, current_stork_pub_key);
}

#[tokio::test]
async fn test_update_stork_single_update_fee_in_wei_owner() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    let new_fee = 2;
    stork_instance.methods().update_single_update_fee_in_wei(new_fee).call().await.unwrap();

    assert_eq!(stork_instance.methods().single_update_fee_in_wei().call().await.unwrap().value, new_fee);
}

#[tokio::test]
async fn test_update_stork_single_update_fee_in_wei_not_owner() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // get current fee
    let current_fee = stork_instance.methods().single_update_fee_in_wei().call().await.unwrap().value;

    let new_fee = 2;
    let res = stork_instance.clone().with_account(not_stork_owner).methods().update_single_update_fee_in_wei(new_fee).call().await;
    assert!(res.is_err());
    assert_eq!(stork_instance.methods().single_update_fee_in_wei().call().await.unwrap().value, current_fee);
}

#[tokio::test]
async fn test_version() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    let version = stork_instance.methods().version().call().await.unwrap().value;
    assert_eq!(version, "1.0.0");
}

#[tokio::test]
async fn test_verify_stork_signature_v1_valid() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // construct evm address
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);

    // construct quantized value - exactly matching the contract test
    let quantized_value_u128 = 62507457175499998000000u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let value_with_indent = quantized_value_u128 + indent;
    let quantized_value = I128 { underlying: value_with_indent };

    let id = Bits256::from_hex_str("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de").unwrap();
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = Bits256::from_hex_str("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741").unwrap();
    let s = Bits256::from_hex_str("0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758").unwrap();
    let v = 28;

    let result = stork_instance.methods().verify_stork_signature_v1(stork_pub_key, id, recv_time, quantized_value, publisher_merkle_root, value_compute_alg_hash, r, s, v).call().await.unwrap();
    println!("result: {:?}", result);
    assert!(result.value);
}

#[tokio::test]
async fn test_verify_stork_signature_v1_invalid() {
    let (stork_owner, _, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // construct evm address
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);

    // construct quantized value - exactly matching the contract test - just a little too high
    let quantized_value_u128 = 62507457175499999000000u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let value_with_indent = quantized_value_u128 + indent;
    let quantized_value = I128 { underlying: value_with_indent };

    let id = Bits256::from_hex_str("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de").unwrap();
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = Bits256::from_hex_str("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741").unwrap();
    let s = Bits256::from_hex_str("0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758").unwrap();
    let v = 28;

    let result = stork_instance.methods().verify_stork_signature_v1(stork_pub_key, id, recv_time, quantized_value, publisher_merkle_root, value_compute_alg_hash, r, s, v).call().await.unwrap();
    assert!(!result.value);
}

#[tokio::test]
async fn test_update_temporal_numeric_value_v1_valid() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // construct quantized value - exactly matching the contract test
    let quantized_value_u128 = 62507457175499998000000u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let value_with_indent = quantized_value_u128 + indent;
    let quantized_value = I128 { underlying: value_with_indent };
    
    let id = Bits256::from_hex_str("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de").unwrap();
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = Bits256::from_hex_str("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741").unwrap();
    let s = Bits256::from_hex_str("0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758").unwrap();
    let v = 28;

    let temporal_numeric_value = TemporalNumericValue {
        timestamp_ns: recv_time,
        quantized_value: quantized_value.clone(),
    };

    let temporal_numeric_value_input = TemporalNumericValueInput {
        temporal_numeric_value,
        id,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    };

    let temporal_numeric_value_input_vec = vec![temporal_numeric_value_input];

    // Get the required fee
    let fee = stork_instance.methods().get_update_fee_v1(temporal_numeric_value_input_vec.clone()).call().await.unwrap().value;

    let tx_policies = TxPolicies::default();
    let call_params = CallParameters::default().with_amount(fee);
    stork_instance.clone()
        .with_account(not_stork_owner)
        .methods()
        .update_temporal_numeric_values_v1(temporal_numeric_value_input_vec)
        .with_tx_policies(tx_policies)
        .call_params(call_params).unwrap()
        .call()
        .await.unwrap();


    // get the temporal numeric value
    let temporal_numeric_value = stork_instance.methods().get_temporal_numeric_value_unchecked_v1(id).call().await.unwrap().value;
    assert_eq!(temporal_numeric_value.timestamp_ns, recv_time);
    assert_eq!(temporal_numeric_value.quantized_value, quantized_value);
}

#[tokio::test]
async fn test_update_temporal_numeric_value_v1_negative() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();
    
    // construct quantized value
    let quantized_value_u128 = 3020199000000u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let quantized_value = I128 { underlying: indent - quantized_value_u128 };

    let id = Bits256::from_hex_str("0x281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf").unwrap();
    let recv_time = 1750794968021348308u64;
    let publisher_merkle_root = Bits256::from_hex_str("0x5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0x14c36cf7272689cec0335efdc5f82dc2d4b1aceb8d2320d3245e4593df32e696").unwrap();
    let s = Bits256::from_hex_str("0x79ab437ecd56dc9fcf850f192328840f7f47d5df57cb939d99146b33014c39f0").unwrap();
    let v = 27;

    let temporal_numeric_value = TemporalNumericValue {
        timestamp_ns: recv_time,
        quantized_value: quantized_value.clone(),
    };

    let temporal_numeric_value_input = TemporalNumericValueInput {
        temporal_numeric_value,
        id,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    };

    let temporal_numeric_value_input_vec = vec![temporal_numeric_value_input];

    // set public key to dev key
    let stork_pub_key_bits = Bits256::from_hex_str("0x0000000000000000000000003db9E960ECfCcb11969509FAB000c0c96DC51830").unwrap();
    let stork_pub_key = EvmAddress::from(stork_pub_key_bits);
    stork_instance.methods().update_stork_public_key(stork_pub_key).call().await.unwrap();

    // Get the required fee
    let fee = stork_instance.methods().get_update_fee_v1(temporal_numeric_value_input_vec.clone()).call().await.unwrap().value;

    let tx_policies = TxPolicies::default();
    let call_params = CallParameters::default().with_amount(fee);

    stork_instance.clone()
        .with_account(not_stork_owner)
        .methods()
        .update_temporal_numeric_values_v1(temporal_numeric_value_input_vec)
        .with_tx_policies(tx_policies)
        .call_params(call_params).unwrap()
        .call()
        .await.unwrap();

    // get the temporal numeric value
    let temporal_numeric_value = stork_instance.methods().get_temporal_numeric_value_unchecked_v1(id).call().await.unwrap().value;
    assert_eq!(temporal_numeric_value.timestamp_ns, recv_time);
    assert_eq!(temporal_numeric_value.quantized_value, quantized_value);
}

#[tokio::test]
async fn test_update_temporal_numeric_value_v1_invalid() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    initialize_stork_default(&stork_instance, stork_owner).await.unwrap();

    // construct quantized value - just a little too high
    let quantized_value_u128 = 62507457175499998000001u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let value_with_indent = quantized_value_u128 + indent;
    let quantized_value = I128 { underlying: value_with_indent };
    
    let id = Bits256::from_hex_str("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de").unwrap();
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = Bits256::from_hex_str("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741").unwrap();
    let s = Bits256::from_hex_str("0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758").unwrap();
    let v = 28;

    let temporal_numeric_value = TemporalNumericValue {
        timestamp_ns: recv_time,
        quantized_value: quantized_value.clone(),
    };

    let temporal_numeric_value_input = TemporalNumericValueInput {
        temporal_numeric_value,
        id,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    };

    let temporal_numeric_value_input_vec = vec![temporal_numeric_value_input];

    // Get the required fee
    let fee = stork_instance.methods().get_update_fee_v1(temporal_numeric_value_input_vec.clone()).call().await.unwrap().value;

    let tx_policies = TxPolicies::default();
    let call_params = CallParameters::default().with_amount(fee);
    
    // This should fail since we modified the value
    let result = stork_instance
        .with_account(not_stork_owner)
        .methods()
        .update_temporal_numeric_values_v1(temporal_numeric_value_input_vec)
        .with_tx_policies(tx_policies)
        .call_params(call_params).unwrap()
        .call()
        .await;

    assert!(result.is_err());
}

#[tokio::test]
async fn test_update_temporal_numeric_values_v1_insufficient_fee() {
    let (stork_owner, not_stork_owner, stork_instance) = setup_tests().await;

    // Set fee to 2 using owner
    stork_instance
        .clone()
        .with_account(stork_owner)
        .methods()
        .update_single_update_fee_in_wei(2)
        .call()
        .await
        .unwrap();

    // Use valid signature values
    let quantized_value_u128 = 62507457175499998000000u128;
    let indent = 1u128 << 127;  // This is I128::indent() in Sway
    let value_with_indent = quantized_value_u128 + indent;
    let quantized_value = I128 { underlying: value_with_indent };
    
    let id = Bits256::from_hex_str("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de").unwrap();
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = Bits256::from_hex_str("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318").unwrap();
    let value_compute_alg_hash = Bits256::from_hex_str("0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba").unwrap();
    let r = Bits256::from_hex_str("0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741").unwrap();
    let s = Bits256::from_hex_str("0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758").unwrap();
    let v = 28;

    let temporal_numeric_value = TemporalNumericValue {
        timestamp_ns: recv_time,
        quantized_value: quantized_value.clone(),
    };

    let temporal_numeric_value_input = TemporalNumericValueInput {
        temporal_numeric_value,
        id,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    };

    let temporal_numeric_value_input_vec = vec![temporal_numeric_value_input];

    let tx_policies = TxPolicies::default();
    // Only send 1 when fee is 2
    let call_params = CallParameters::default().with_amount(1);
    
    let result = stork_instance
        .with_account(not_stork_owner)
        .methods()
        .update_temporal_numeric_values_v1(temporal_numeric_value_input_vec)
        .with_tx_policies(tx_policies)
        .call_params(call_params).unwrap()
        .call()
        .await;

    assert!(result.is_err());
}



