// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package chain_pusher

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// TransferOwnership is the `transfer_ownership` instruction.
type TransferOwnership struct {
	NewOwner *ag_solanago.PublicKey

	// [0] = [WRITE] config
	//
	// [1] = [SIGNER] owner
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewTransferOwnershipInstructionBuilder creates a new `TransferOwnership` instruction builder.
func NewTransferOwnershipInstructionBuilder() *TransferOwnership {
	nd := &TransferOwnership{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetNewOwner sets the "new_owner" parameter.
func (inst *TransferOwnership) SetNewOwner(new_owner ag_solanago.PublicKey) *TransferOwnership {
	inst.NewOwner = &new_owner
	return inst
}

// SetConfigAccount sets the "config" account.
func (inst *TransferOwnership) SetConfigAccount(config ag_solanago.PublicKey) *TransferOwnership {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(config).WRITE()
	return inst
}

func (inst *TransferOwnership) findFindConfigAddress(knownBumpSeed uint8) (pda ag_solanago.PublicKey, bumpSeed uint8, err error) {
	var seeds [][]byte
	// const: stork_config
	seeds = append(seeds, []byte{byte(0x73), byte(0x74), byte(0x6f), byte(0x72), byte(0x6b), byte(0x5f), byte(0x63), byte(0x6f), byte(0x6e), byte(0x66), byte(0x69), byte(0x67)})

	if knownBumpSeed != 0 {
		seeds = append(seeds, []byte{byte(bumpSeed)})
		pda, err = ag_solanago.CreateProgramAddress(seeds, ProgramID)
	} else {
		pda, bumpSeed, err = ag_solanago.FindProgramAddress(seeds, ProgramID)
	}
	return
}

// FindConfigAddressWithBumpSeed calculates Config account address with given seeds and a known bump seed.
func (inst *TransferOwnership) FindConfigAddressWithBumpSeed(bumpSeed uint8) (pda ag_solanago.PublicKey, err error) {
	pda, _, err = inst.findFindConfigAddress(bumpSeed)
	return
}

func (inst *TransferOwnership) MustFindConfigAddressWithBumpSeed(bumpSeed uint8) (pda ag_solanago.PublicKey) {
	pda, _, err := inst.findFindConfigAddress(bumpSeed)
	if err != nil {
		panic(err)
	}
	return
}

// FindConfigAddress finds Config account address with given seeds.
func (inst *TransferOwnership) FindConfigAddress() (pda ag_solanago.PublicKey, bumpSeed uint8, err error) {
	pda, bumpSeed, err = inst.findFindConfigAddress(0)
	return
}

func (inst *TransferOwnership) MustFindConfigAddress() (pda ag_solanago.PublicKey) {
	pda, _, err := inst.findFindConfigAddress(0)
	if err != nil {
		panic(err)
	}
	return
}

// GetConfigAccount gets the "config" account.
func (inst *TransferOwnership) GetConfigAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOwnerAccount sets the "owner" account.
func (inst *TransferOwnership) SetOwnerAccount(owner ag_solanago.PublicKey) *TransferOwnership {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *TransferOwnership) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst TransferOwnership) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_TransferOwnership,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst TransferOwnership) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TransferOwnership) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.NewOwner == nil {
			return errors.New("NewOwner parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Config is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Owner is not set")
		}
	}
	return nil
}

func (inst *TransferOwnership) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("TransferOwnership")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param(" NewOwner", *inst.NewOwner))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("config", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta(" owner", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj TransferOwnership) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `NewOwner` param:
	err = encoder.Encode(obj.NewOwner)
	if err != nil {
		return err
	}
	return nil
}
func (obj *TransferOwnership) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `NewOwner`:
	err = decoder.Decode(&obj.NewOwner)
	if err != nil {
		return err
	}
	return nil
}

// NewTransferOwnershipInstruction declares a new TransferOwnership instruction with the provided parameters and accounts.
func NewTransferOwnershipInstruction(
	// Parameters:
	new_owner ag_solanago.PublicKey,
	// Accounts:
	config ag_solanago.PublicKey,
	owner ag_solanago.PublicKey) *TransferOwnership {
	return NewTransferOwnershipInstructionBuilder().
		SetNewOwner(new_owner).
		SetConfigAccount(config).
		SetOwnerAccount(owner)
}
