#include <stdint.h>

int hash_and_sign(char *asset, char *quantized_price, const int64_t *timestamp_ns, const unsigned char *oracle_name_int_ptr, const unsigned char *pk_ptr,
                  unsigned char *pedersen_hash_ptr, unsigned char *sig_r_ptr, unsigned char *sig_s_ptr);