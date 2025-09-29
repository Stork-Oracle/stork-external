;; Minimal Stork contract for local development using free namespace
(namespace 'free)

(module stork-test2 GOVERNANCE
    ;; Imports
    (use coin)

    ;; Constants
    (defconst STATE_KEY:string "state")
    (defconst ADMIN_ACCOUNT:string "k:5cc0092889287113bd1d44beebc4f57ae8c46c915702df7870fdac83aae27e4d")

    ; Return constants
    (defconst INITIALIZED_CONTRACT:string "Successfully initialized contract") 

    ;; Capabilities
    (defcap GOVERNANCE () 
        @doc "Governance capability for the module."
        true
    )    

    ;; State storage
    (defschema state
        @doc "Schema for the state of the contract."
        storkEvmPublicKey:string
        singleUpdateFeeInStu:integer
    )

    (deftable state-table:{state})

    ;; Admin Functions
    (defun initialize:string (storkEvmPublicKey:string singleUpdateFeeInStu:integer)
        @doc "Initializes the contract with the given stork EVM public key and single update fee in stu."

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

    (defun get-stork-evm-public-key:string ()
        @doc "Gets the stork EVM public key from the state."
        
        (with-read
            state-table
            STATE_KEY
            {
                "storkEvmPublicKey" := storkEvmPublicKey 
            }
            storkEvmPublicKey 
        )
    )

    (defun get-single-update-fee-in-stu:integer ()
        @doc "Gets the single update fee in stu from the state."
        
        (with-read state-table STATE_KEY
        {
            "singleUpdateFeeInStu" := singleUpdateFeeInStu
            }
            singleUpdateFeeInStu
        )
    )

    (defun is-initialized:bool ()
        @doc "Checks whether the contract is initialized."
        
        (contains STATE_KEY (keys state-table))
    )
)

; Create table
(create-table state-table)