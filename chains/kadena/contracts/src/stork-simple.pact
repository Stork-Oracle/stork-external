;; Simplified Stork contract without namespace for local development
(define-keyset "stork-admin" (read-keyset "stork"))

(module stork GOVERNANCE
    ;; Imports
    (use coin)

    ;; Constants
    (defconst STATE_KEY:string "state")

    ; Error constants
    (defconst ERR_FEED_NOT_FOUND:string "Feed not found")
    (defconst ERR_NOT_INITIALIZED:string "Contract is not initialized")
    (defconst ERR_ALREADY_INITIALIZED:string "Contract is already initialized")
    (defconst ERR_INSUFFICIENT_FEE:string "Insufficient fee")

    ; Return constants
    (defconst UPDATED_TEMPORAL_NUMERIC_VALUE:string "Successfully updated temporal numeric value")
    (defconst NOT_UPDATED_TEMPORAL_NUMERIC_VALUE:string "Did not update temporal numeric value; value is stale")
    (defconst UPDATED_SINGLE_UPDATE_FEE_IN_STU:string "Successfully updated single update fee in stu")
    (defconst UPDATED_STORK_EVM_PUBLIC_KEY:string "Successfully updated stork EVM public key")
    (defconst INITIALIZED_CONTRACT:string "Successfully initialized contract") 

    ; Treasury
    (defconst TREASURY_ACCOUNT:string "stork-treasury")

    ;; Capabilities

    (defcap GOVERNANCE () 
        @doc "Governance capability for the module."
        (enforce-keyset "stork-admin")
    )    

    (defcap INITIALIZED ()
        @doc "Used to enforce that the contract is initialized."
        (let ((initialized (is-initialized)))
            (enforce initialized ERR_NOT_INITIALIZED)
            true 
        )
    )

    (defcap NOT_INITIALIZED ()
        @doc "Used to enforce that the contract is not initialized."
        (let ((initialized (is-initialized)))
            (enforce (not initialized) ERR_ALREADY_INITIALIZED)
            true
        )
    )

    (defcap TEMPORAL_NUMERIC_VALUE_EXISTS (encodedAssetId:string)
        @doc "Used to enforce that a temporal numeric value exists for a specific encoded asset id."
        (let ((exists (temporal-numeric-value-exists encodedAssetId)))
            (enforce exists ERR_FEED_NOT_FOUND)
            true
        )
    )

    ;; Events
    
    (defcap CONTRACT_INITIALIZED (storkEvmPublicKey:string singleUpdateFeeInStu:integer)
        @doc "Event emitted when the contract is initialized."
        @event 
        true
    )

    (defcap STORK_EVM_PUBLIC_KEY_UPDATED (newStorkEvmPublicKey:string)
        @doc "Event emitted when the stork evm public key on the state is updated."
        @event
        true
    )

    (defcap SINGLE_UPDATE_FEE_IN_STU_UPDATED (newSinglUpdateFeeInStu:integer)
        @doc "Event emitted when the single update fee in stu on the state is updated."
        @event
        true
    )

    (defcap VALUE_UPDATE (encodedAssetId:string temporalNumericValue:object{temporal-numeric-value})
        @doc "Event emitted when a temporal numeric value is updated."
        @event
        true
    )

    ;; State storage.
    ;; We use a table for state storage, but as an invariant of this contract there's only ever one entry in the table.
    ;; The single entry is keyed by the STATE_KEY const.

    ; schema 
    (defschema state
        @doc "Schema for the state of the contract. \
        \ - storkEvmPublicKey:string - The stork EVM public key. \
        \ - singleUpdateFeeInStu:integer - The single update fee in stu. \
        \ Note: This table is intended to be used as a single entry table."
        storkEvmPublicKey:string
        singleUpdateFeeInStu:integer
    )

    ; table 
    (deftable state-table:{state})

    ;; Storage for temporal numeric values

    ; Schema
    (defschema temporal-numeric-value
        @doc "Schema for a temporal numeric value. Intended use with a table is to use the encoded asset id as the key. \
        \ - timestampNs:integer - The unix nanosecond timestamp of the temporal numeric value. \
        \ - quantizedValue:integer - The quantized value of the temporal numeric value."
        timestampNs:integer
        quantizedValue:integer
    )
    
    ; Table
    (deftable temporal-numeric-values-table:{temporal-numeric-value})

    ;; Admin Functions

    (defun initialize:string (storkEvmPublicKey:string singleUpdateFeeInStu:integer)
        @doc "Initializes the contract with the given stork EVM public key and single update fee in stu. \
        \ Parameters: \
        \ - storkEvmPublicKey:string - The stork EVM public key. \
        \ - singleUpdateFeeInStu:integer - The single update fee in stu. \
        \ Returns \
        \ string - Success message "

        (with-capability (NOT_INITIALIZED)
            (with-capability (GOVERNANCE)
                ; Set up state
                (insert state-table STATE_KEY
                    {
                        "storkEvmPublicKey": storkEvmPublicKey,
                        "singleUpdateFeeInStu": singleUpdateFeeInStu
                    }
                )
                INITIALIZED_CONTRACT
            )
        )
    )

    (defun update-stork-evm-public-key (storkEvmPublicKey:string)
        @doc "Updates the stork EVM public key in the state. \
        \ Parameters: \
        \ - string - The stork EVM public key. \
        \ Returns \
        \ string - Success message "

        (with-capability (INITIALIZED)
            (with-capability (GOVERNANCE)
                (update state-table STATE_KEY
                    {
                        "storkEvmPublicKey": storkEvmPublicKey,
                        "singleUpdateFeeInStu": (get-single-update-fee-in-stu)
                    }
                )
                UPDATED_STORK_EVM_PUBLIC_KEY
            )
        )
    )

    (defun update-single-update-fee-in-stu (singleUpdateFeeInStu:integer)
        @doc "Updates the single update fee in stu in the state. \
        \ Parameters: \
        \ - singleUpdateFeeInStu:integer - The single update fee in stu. \
        \ Returns \
        \ string - Success message "

        (with-capability (INITIALIZED)
            (with-capability (GOVERNANCE)
                (update state-table STATE_KEY
                    {
                        "storkEvmPublicKey": (get-stork-evm-public-key),
                        "singleUpdateFeeInStu": singleUpdateFeeInStu
                    }
                )
                UPDATED_SINGLE_UPDATE_FEE_IN_STU
            )
        )
    )

    ;; State Getters

    (defun get-stork-evm-public-key:string ()
        @doc "Gets the stork EVM public key from the state. \
        \ Returns: \
        \ - string - The stork EVM public key."
        
        (with-capability (INITIALIZED)
            (with-read
                state-table
                STATE_KEY
                {
                    "storkEvmPublicKey" := storkEvmPublicKey 
                }
                storkEvmPublicKey 
            )
        )
    )

    (defun get-single-update-fee-in-stu:integer ()
        @doc "Gets the single update fee in stu from the state. \
        \ Returns: \
        \ - integer - The single update fee in stu."
        
        (with-capability (INITIALIZED)
            (with-read state-table STATE_KEY
            {
                "singleUpdateFeeInStu" := singleUpdateFeeInStu
                }
                singleUpdateFeeInStu
            )
        )
    )

    ;; Helper functions

    (defun is-initialized:bool ()
        @doc "Checks whether the contract is initialized. \
        \ Returns: \
        \ - bool - True if the contract is initialized, false otherwise."
        
        (contains STATE_KEY (keys state-table))
    )

    (defun temporal-numeric-value-exists:bool (encodedAssetId:string)
        @doc "Checks whether a temporal numeric value exists for a specific encoded asset id. \
        \ Parameters: \
        \ - encodedAssetId:string - The encoded asset id in the form of a hex string. \
        \ Returns: \
        \ - bool - True if the temporal numeric value exists, false otherwise."
        
        ; requiring this capability without attempting to acquire effectively makes this an internal function
        (require-capability (INITIALIZED))
        (contains encodedAssetId (keys temporal-numeric-values-table))
    )
)
; Create tables necessary for the contract
(create-table temporal-numeric-values-table)
(create-table state-table)