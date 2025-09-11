use crate::contract::UpdateData;
use crate::contract::{
    sv::mt::{CodeId, StorkContractProxy},
    StorkContract,
};
use crate::temporal_numeric_value::TemporalNumericValue;
use sylvia::cw_multi_test::IntoAddr;
use sylvia::cw_std::Coin;
use sylvia::multitest::{App, Proxy};

const STORK_EVM_PUBLIC_KEY: &str = "0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const SINGLE_UPDATE_FEE: u128 = 2;
const OWNER: &str = "owner";
const USER: &str = "user";
const UNDERFUNDED_USER: &str = "underfunded_user";

pub fn hex_to_bytes(hex: &str) -> Vec<u8> {
    let hex = hex.trim_start_matches("0x");
    (0..hex.len())
        .step_by(2)
        .map(|i| u8::from_str_radix(&hex[i..i + 2], 16).unwrap())
        .collect()
}

fn instantiate<'a>(
    app: &'a mut App<sylvia::cw_multi_test::App>,
) -> Proxy<
    'a,
    sylvia::cw_multi_test::App,
    StorkContract<sylvia::cw_std::Empty, sylvia::cw_std::Empty>,
> {
    let stork_evm_public_key = hex_to_bytes(STORK_EVM_PUBLIC_KEY).try_into().unwrap();
    let single_update_fee = Coin::new(SINGLE_UPDATE_FEE, "stork");

    app.app_mut().init_modules(|router, _, storage| {
        router
            .bank
            .init_balance(
                storage,
                &OWNER.into_addr(),
                vec![Coin::new(100u128, "stork")],
            )
            .unwrap();
        router
            .bank
            .init_balance(
                storage,
                &USER.into_addr(),
                vec![Coin::new(100u128, "stork")],
            )
            .unwrap();
        router
            .bank
            .init_balance(
                storage,
                &UNDERFUNDED_USER.into_addr(),
                vec![Coin::new(1u128, "stork")],
            )
            .unwrap();
    });

    let code_id = CodeId::store_code(app);

    let contract = code_id
        .instantiate(stork_evm_public_key, single_update_fee)
        .with_label("StorkContract")
        .with_admin(Some(OWNER))
        .call(&OWNER.into_addr())
        .unwrap();

    contract
}

#[test]
fn test_instantiate() {
    let mut app = App::default();
    instantiate(&mut app);
}

#[test]
fn test_get_stork_evm_public_key() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let response = contract.get_stork_evm_public_key();
    let pubkey: [u8; 20] = hex_to_bytes(STORK_EVM_PUBLIC_KEY).try_into().unwrap();
    assert_eq!(response.unwrap().stork_evm_public_key, pubkey);
}

#[test]
fn test_get_single_update_fee() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let response = contract.get_single_update_fee();
    assert_eq!(response.unwrap().fee, Coin::new(SINGLE_UPDATE_FEE, "stork"));
}

#[test]
fn test_get_owner() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let response = contract.get_owner();
    assert_eq!(response.unwrap().owner, OWNER.into_addr());
}

fn update_temporal_numeric_value_helper(
    contract: &mut Proxy<
        '_,
        sylvia::cw_multi_test::App,
        StorkContract<sylvia::cw_std::Empty, sylvia::cw_std::Empty>,
    >,
    user: &str,
    update_fee: u128,
    id_hex: &str,
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root_hex: &str,
    value_compute_alg_hash_hex: &str,
    r_hex: &str,
    s_hex: &str,
    v: u8,
) -> UpdateData {
    let id = hex_to_bytes(id_hex)[..32]
        .try_into()
        .unwrap();
    let recv_time: u64 = recv_time;
    let quantized_value: i128 = quantized_value;
    let publisher_merkle_root =
        hex_to_bytes(publisher_merkle_root_hex)[..32]
            .try_into()
            .unwrap();
    let value_compute_alg_hash =
        hex_to_bytes(value_compute_alg_hash_hex)[..32]
            .try_into()
            .unwrap();
    let r = hex_to_bytes(r_hex)[..32]
        .try_into()
        .unwrap();
    let s = hex_to_bytes(s_hex)[..32]
        .try_into()
        .unwrap();
    let v = v;

    let temporal_numeric_value = TemporalNumericValue {
        timestamp_ns: recv_time.into(),
        quantized_value: quantized_value.into(),
    };

    let update_data = UpdateData {
        id,
        temporal_numeric_value,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    };

    let update_data_vec = vec![update_data.clone()];
    if let Err(e) = contract
        .update_temporal_numeric_values_evm(update_data_vec)
        .with_funds(&[Coin::new(update_fee, "stork")])
        .call(&user.into_addr())
    {
        panic!("{}", e);
    }
    update_data
}

#[test]
fn test_update_temporal_numeric_value_and_get_value() {
    let mut app = App::default();
    let mut contract = instantiate(&mut app);
    let update_data = update_temporal_numeric_value_helper(&mut contract, USER, SINGLE_UPDATE_FEE, "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de", 1722632569208762117, 62507457175499998000000, "e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318", "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba", "b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741", "16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758", 28);
    let response = contract.get_latest_canonical_temporal_numeric_value_unchecked(update_data.id);
    assert_eq!(
        response.unwrap().temporal_numeric_value,
        update_data.temporal_numeric_value
    );
}

#[test]
fn test_update_temporal_numeric_value_negative_value() {
    let mut app = App::default();
    let mut contract = instantiate(&mut app);

    // set to dev key
    let pubkey: [u8; 20] = hex_to_bytes("3db9E960ECfCcb11969509FAB000c0c96DC51830")
        .try_into()
        .unwrap();
    contract
        .set_stork_evm_public_key(pubkey)
        .call(&OWNER.into_addr())
        .unwrap();

    let update_data = update_temporal_numeric_value_helper(&mut contract, USER, SINGLE_UPDATE_FEE, "281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf", 1750794968021348308, -3020199000000, "5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878", "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba", "14c36cf7272689cec0335efdc5f82dc2d4b1aceb8d2320d3245e4593df32e696", "79ab437ecd56dc9fcf850f192328840f7f47d5df57cb939d99146b33014c39f0", 27);

    let response = contract.get_latest_canonical_temporal_numeric_value_unchecked(update_data.id);
    assert_eq!(
        response.unwrap().temporal_numeric_value,
        update_data.temporal_numeric_value
    );
}
#[test]
#[should_panic(expected = "Feed not found")]
fn test_get_latest_canonical_temporal_numeric_value_unchecked_invalid_id() {
    let mut app = App::default();
    let mut contract = instantiate(&mut app);
    let update_data = update_temporal_numeric_value_helper(&mut contract, USER, SINGLE_UPDATE_FEE, "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de", 1722632569208762117, 62507457175499998000000, "e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318", "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba", "b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741", "16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758", 28);
    let id = hex_to_bytes("0000000000000000000000000000000000000000000000000000000000000000")[..32]
        .try_into()
        .unwrap();
    let response = contract.get_latest_canonical_temporal_numeric_value_unchecked(id);
    assert_eq!(
        response.unwrap().temporal_numeric_value,
        update_data.temporal_numeric_value
    );
}

#[test]
#[should_panic(expected = "Insufficient funds")]
fn test_update_temporal_numeric_value_insufficient_funds() {
    let mut app = App::default();
    let mut contract = instantiate(&mut app);
    update_temporal_numeric_value_helper(&mut contract, UNDERFUNDED_USER, 1, "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de", 1722632569208762117, 62507457175499998000000, "e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318", "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba", "b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741", "16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758", 28);
}

#[test]
fn test_set_owner() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    contract
        .set_owner(USER.into_addr())
        .call(&OWNER.into_addr())
        .unwrap();
    let response = contract.get_owner();
    assert_eq!(response.unwrap().owner, USER.into_addr());
}

#[test]
#[should_panic(expected = "Not Authorized")]
fn test_set_owner_not_authorized() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    if let Err(e) = contract.set_owner(USER.into_addr()).call(&USER.into_addr()) {
        panic!("{}", e);
    }
}

#[test]
fn test_set_stork_evm_public_key() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let pubkey: [u8; 20] = hex_to_bytes("0000000000000000000000000000000000000000")
        .try_into()
        .unwrap();
    contract
        .set_stork_evm_public_key(pubkey)
        .call(&OWNER.into_addr())
        .unwrap();
    let response = contract.get_stork_evm_public_key();
    assert_eq!(response.unwrap().stork_evm_public_key, pubkey);
}

#[test]
#[should_panic(expected = "Not Authorized")]
fn test_set_stork_evm_public_key_not_authorized() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let pubkey: [u8; 20] = hex_to_bytes(STORK_EVM_PUBLIC_KEY).try_into().unwrap();
    if let Err(e) = contract
        .set_stork_evm_public_key(pubkey)
        .call(&USER.into_addr())
    {
        panic!("{}", e);
    }
}

#[test]
fn test_set_single_update_fee() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let fee = Coin::new(100000u128, "stork");
    contract
        .set_single_update_fee(fee.clone())
        .call(&OWNER.into_addr())
        .unwrap();
    let response = contract.get_single_update_fee();
    assert_eq!(response.unwrap().fee, fee);
}

#[test]
#[should_panic(expected = "Not Authorized")]
fn test_set_single_update_fee_not_authorized() {
    let mut app = App::default();
    let contract = instantiate(&mut app);
    let fee = Coin::new(100000u128, "stork");
    if let Err(e) = contract
        .set_single_update_fee(fee.clone())
        .call(&USER.into_addr())
    {
        panic!("{}", e);
    }
}
