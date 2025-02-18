/**
* This file was automatically generated by @cosmwasm/ts-codegen@1.12.0.
* DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
* and run the @cosmwasm/ts-codegen generate command to regenerate this file.
*/

import { CosmWasmClient, SigningCosmWasmClient, ExecuteResult } from "@cosmjs/cosmwasm-stargate";
import { Coin, StdFee } from "@cosmjs/amino";
import { Addr, InstantiateMsg, ExecuteMsg, ExecMsg, QueryMsg, QueryMsg1 } from "./Example.types";
export interface ExampleReadOnlyInterface {
  contractAddress: string;
}
export class ExampleQueryClient implements ExampleReadOnlyInterface {
  client: CosmWasmClient;
  contractAddress: string;
  constructor(client: CosmWasmClient, contractAddress: string) {
    this.client = client;
    this.contractAddress = contractAddress;
  }
}
export interface ExampleInterface extends ExampleReadOnlyInterface {
  contractAddress: string;
  sender: string;
  useStorkPrice: ({
    feedId
  }: {
    feedId: number[];
  }, fee_?: number | StdFee | "auto", memo_?: string, funds_?: Coin[]) => Promise<ExecuteResult>;
}
export class ExampleClient extends ExampleQueryClient implements ExampleInterface {
  client: SigningCosmWasmClient;
  sender: string;
  contractAddress: string;
  constructor(client: SigningCosmWasmClient, sender: string, contractAddress: string) {
    super(client, contractAddress);
    this.client = client;
    this.sender = sender;
    this.contractAddress = contractAddress;
    this.useStorkPrice = this.useStorkPrice.bind(this);
  }
  useStorkPrice = async ({
    feedId
  }: {
    feedId: number[];
  }, fee_: number | StdFee | "auto" = "auto", memo_?: string, funds_?: Coin[]): Promise<ExecuteResult> => {
    return await this.client.execute(this.sender, this.contractAddress, {
      use_stork_price: {
        feed_id: feedId
      }
    }, fee_, memo_, funds_);
  };
}