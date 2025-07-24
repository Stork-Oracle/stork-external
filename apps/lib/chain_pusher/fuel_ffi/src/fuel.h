#ifndef FUEL_H
#define FUEL_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct FuelClient FuelClient;

// Client management
FuelClient* fuel_client_new(const char* config_json);
void fuel_client_free(FuelClient* client);

// Contract interactions
char* fuel_get_latest_value(FuelClient* client, const uint8_t* id);
char* fuel_update_values(FuelClient* client, const char* inputs_json);
uint64_t fuel_get_wallet_balance(FuelClient* client);
char* fuel_get_last_error();

// Memory management
void fuel_free_string(char* s);

#ifdef __cplusplus
}
#endif

#endif // FUEL_H