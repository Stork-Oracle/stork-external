#include <stdint.h>

int hash_and_sign(const unsigned char *x_ptr, const unsigned char *y_ptr, const unsigned char *pk_ptr,
                  unsigned char *pedersen_hash_ptr, unsigned char *sig_r_ptr, unsigned char *sig_s_ptr);