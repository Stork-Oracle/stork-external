use abigen_bindings::stork_proxy_mod::standards;
use fuels::{
    core::{
        codec::EncoderConfig,
        traits::{Parameterize, Tokenizable},
    },
    prelude::*,
    programs::responses::CallResponse,
    types::{Identity, Token},
};

macro_rules! tokenize {
    ($($x:expr),* $(,)?) => {
        &[$($x.into_token())*]
    }
}

abigen! {
    Contract(
        name = "Stork",
        abi = "stork/out/debug/stork-abi.json"
    ),
    Contract(
        name = "StorkProxy",
        abi = "stork_proxy/out/debug/stork_proxy-abi.json"
    )
}

struct StorkSetup {
    pub wallets: Vec<Wallet>,
    pub stork_id: ContractId,
    pub stork_proxy: StorkProxy<Wallet>,
    pub stork_proxy_id: ContractId,
}

impl StorkSetup {
    pub async fn new() -> Result<Self> {
        let wallets = launch_custom_provider_and_get_wallets(
            WalletsConfig::new(
                Some(3),             // Create 3 wallets for Deployer, StorkPubKey and Tester
                Some(1),             /* Single coin (UTXO) */
                Some(1_000_000_000), /* Amount per coin */
            ),
            None,
            None,
        )
        .await?;

        let deployer = &wallets[0];

        let stork_contract =
            Contract::load_from("../stork/out/debug/stork.bin", LoadConfiguration::default())?;
        let stork_storage = stork_contract.storage_slots().to_vec();

        let stork_id = stork_contract
            .convert_to_loader(100)?
            .deploy(deployer, TxPolicies::default())
            .await?;

        let stork_proxy_contract =
            Contract::load_from("./out/debug/stork_proxy.bin", LoadConfiguration::default())?;

        let combined_storage = [&stork_storage, stork_proxy_contract.storage_slots()].concat();

        let stork_proxy_id = stork_proxy_contract
            .with_storage_slots(combined_storage)
            .deploy(deployer, TxPolicies::default())
            .await?;

        let stork_proxy = StorkProxy::new(stork_proxy_id.contract_id.clone(), deployer.clone());

        let _ = stork_proxy
            .methods()
            .set_owner(deployer.address().into())
            .call()
            .await?;

        let _ = stork_proxy
            .methods()
            .set_proxy_target(stork_id.contract_id.clone())
            .call()
            .await?;

        let result = Self {
            wallets,
            stork_id: stork_id.contract_id.into(),
            stork_proxy: stork_proxy,
            stork_proxy_id: stork_proxy_id.contract_id.clone().into(),
        };

        let _ = result
            .call_function::<()>(
                stork_proxy_id.contract_id.into(),
                result.get_deployer_wallet(),
                "initialize",
                tokenize![
                    Identity::Address(result.get_deployer_wallet().address().into()),
                    Identity::Address(result.get_stork_public_key_wallet().address().into()),
                    60u64,
                    1u64,
                ],
            )
            .await?;

        Ok(result)
    }

    pub async fn call_function<T: Tokenizable + Parameterize + std::fmt::Debug>(
        &self,
        contract: ContractId,
        account: &Wallet,
        name: &str,
        args: &[Token],
    ) -> Result<CallResponse<T>> {
        CallHandler::new_contract_call(
            contract.into(),
            account.clone(),
            fuels::core::codec::encode_fn_selector(name),
            args,
            self.stork_proxy.log_decoder().clone(),
            false,
            EncoderConfig::default(),
        )
        .with_contract_ids(&[self.stork_id.into()])
        .call()
        .await
    }

    pub fn get_deployer_wallet(&self) -> &Wallet {
        &self.wallets[0]
    }

    pub fn get_stork_public_key_wallet(&self) -> &Wallet {
        &self.wallets[1]
    }

    pub fn get_tester_wallet(&self) -> &Wallet {
        &self.wallets[2]
    }
}

#[inline(always)]
fn create_identity_token(identity: Identity) -> Result<Token> {
    Ok(Token::Struct(match identity {
        Identity::Address(address) => vec![Token::U64(0), Token::B256(*address)],
        Identity::ContractId(contract_id) => vec![Token::U64(1), Token::B256(*contract_id)],
    }))
}

#[tokio::test]
async fn test_deploy() -> Result<()> {
    let stork_setup = StorkSetup::new().await?;

    let stork_owner = stork_setup
        .call_function(
            stork_setup.stork_proxy_id,
            stork_setup.get_deployer_wallet(),
            "owner",
            &[],
        )
        .await?;
    let stork_proxy_owner = stork_setup
        .stork_proxy
        .methods()
        .proxy_owner()
        .call()
        .await?;

    match (stork_owner.value, stork_proxy_owner.value) {
        (State::Uninitialized, State::Uninitialized) => {}
        (State::Initialized(stork_owner), State::Initialized(stork_proxy_owner)) => {
            assert!(stork_owner == stork_proxy_owner);
        }
        (State::Revoked, State::Revoked) => {}

        (a, b) => {
            panic!("Mismatched Owner State {a:#?} - {b:#?}");
        }
    }

    assert!(stork_setup.stork_id != ContractId::zeroed());
    assert!(stork_setup.stork_proxy_id != ContractId::zeroed());
    Ok(())
}

#[tokio::test]
async fn test_proxy_owner() -> Result<()> {
    let stork_setup = StorkSetup::new().await?;

    let result = stork_setup
        .stork_proxy
        .clone()
        .with_account(stork_setup.get_tester_wallet().clone())
        .methods()
        .set_proxy_target(stork_setup.stork_id)
        .call()
        .await;

    assert!(result.is_err());

    if let Err(e) = result {
        assert!(e.to_string().contains("NotOwner"));
    } else {
        panic!("Wrong Error Message : {:#?}", result);
    }

    Ok(())
}

#[tokio::test]
async fn test_stork_initialize() -> Result<()> {
    let stork_setup = StorkSetup::new().await?;

    let result = stork_setup
        .call_function::<()>(
            stork_setup.stork_proxy_id,
            stork_setup.get_deployer_wallet(),
            "initialize",
            tokenize![
                Identity::Address(stork_setup.get_deployer_wallet().address().into()),
                Identity::Address(stork_setup.get_stork_public_key_wallet().address().into()),
                100u64,
                100u64,
            ],
        )
        .await;
    assert!(result.is_err());

    if let Err(e) = result {
        assert!(e.to_string().contains("Already initialized"));
    } else {
        panic!("Wrong Error Message : {:#?}", result);
    }

    Ok(())
}

#[tokio::test]
async fn test_stork_only_owner() -> Result<()> {
    let stork_setup = StorkSetup::new().await?;

    let result = stork_setup
        .call_function::<standards::src5::State>(
            stork_setup.stork_proxy_id,
            stork_setup.get_tester_wallet(),
            "update_valid_time_period_seconds",
            tokenize![0u64,],
        )
        .await;

    assert!(result.is_err());

    if let Err(e) = result {
        assert!(e.to_string().contains("Only Owner"));
    } else {
        panic!("Wrong Error Message : {:#?}", result);
    }

    let result = stork_setup
        .call_function::<standards::src5::State>(
            stork_setup.stork_proxy_id,
            stork_setup.get_tester_wallet(),
            "update_single_update_fee_in_wei",
            tokenize![100u64,],
        )
        .await;

    assert!(result.is_err());

    if let Err(e) = result {
        assert!(e.to_string().contains("Only Owner"));
    } else {
        panic!("Wrong Error Message : {:#?}", result);
    }

    let result = stork_setup
        .call_function::<standards::src5::State>(
            stork_setup.stork_proxy_id,
            stork_setup.get_tester_wallet(),
            "update_stork_public_key",
            tokenize![Identity::from(stork_setup.get_tester_wallet().address()),],
        )
        .await;

    assert!(result.is_err());

    if let Err(e) = result {
        assert!(e.to_string().contains("Only Owner"));
    } else {
        panic!("Wrong Error Message : {:#?}", result);
    }

    Ok(())
}

#[tokio::test]
async fn test_stork_owner() -> Result<()> {
    let stork_setup = StorkSetup::new().await?;

    let stork_owner = stork_setup
        .call_function::<standards::src5::State>(
            stork_setup.stork_proxy_id,
            stork_setup.get_deployer_wallet(),
            "owner",
            &[],
        )
        .await?;

    let standards::src5::State::Initialized(stork_owner) = stork_owner.value else {
        panic!()
    };
    assert!(stork_owner == stork_setup.get_deployer_wallet().address().into());

    Ok(())
}
