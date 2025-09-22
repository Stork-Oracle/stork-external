(namespace 'stork)

(module stork GOVERNANCE
    (defcap GOVERNANCE() true)    

    (defun returnPhrase (a b )
        (format "My {} has {}" [a b]) 
    )
)
