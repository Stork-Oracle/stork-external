#include <stdint.h>

int hash_and_sign(const unsigned char *x_ptr, const unsigned char *y_ptr, const unsigned char *pk_ptr,
                  unsigned char *pedersen_hash_ptr, unsigned char *sig_r_ptr, unsigned char *sig_s_ptr);

int validate_stark_signature(const unsigned char *x_ptr, const unsigned char *y_ptr, const unsigned char *public_key_ptr,
                  const unsigned char *sig_r_ptr, const unsigned char *sig_s_ptr);
