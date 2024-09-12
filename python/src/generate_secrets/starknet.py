# Pulled from different parts of the Starkware signature module
# on github: https://github.com/starkware-libs/starkex-resources/tree/master/crypto/starkware/crypto/signature

import json
import os
import secrets
from sympy.core.numbers import igcdex
from typing import Tuple

ECPoint = Tuple[int, int]

PEDERSEN_HASH_POINT_FILENAME = os.path.join(
    os.path.dirname(__file__), 'pedersen_params.json')

PEDERSEN_PARAMS = json.load(open(PEDERSEN_HASH_POINT_FILENAME))

EC_ORDER = PEDERSEN_PARAMS['EC_ORDER']
CONSTANT_POINTS = PEDERSEN_PARAMS['CONSTANT_POINTS']
EC_GEN = CONSTANT_POINTS[1]
ALPHA = PEDERSEN_PARAMS['ALPHA']
FIELD_PRIME = PEDERSEN_PARAMS['FIELD_PRIME']

assert EC_GEN == [
    0x1ef15c18599971b7beced415a40f0c7deacfd9b0d1819e03d723d8bc943cfca,
    0x5668060aa49730b7be4801df46ec62de53ecd11abe43a32873000c36e8dc1f]


def div_mod(n: int, m: int, p: int) -> int:
    """
    Finds a nonnegative integer 0 <= x < p such that (m * x) % p == n
    """
    a, b, c = igcdex(m, p)
    assert c == 1
    return (n * a) % p


def ec_add(point1: ECPoint, point2: ECPoint, p: int) -> ECPoint:
    """
    Gets two points on an elliptic curve mod p and returns their sum.
    Assumes the points are given in affine form (x, y) and have different x coordinates.
    """
    assert (point1[0] - point2[0]) % p != 0
    m = div_mod(point1[1] - point2[1], point1[0] - point2[0], p)
    x = (m * m - point1[0] - point2[0]) % p
    y = (m * (point1[0] - x) - point1[1]) % p
    return x, y


def ec_double(point: ECPoint, alpha: int, p: int) -> ECPoint:
    """
    Doubles a point on an elliptic curve with the equation y^2 = x^3 + alpha*x + beta mod p.
    Assumes the point is given in affine form (x, y) and has y != 0.
    """
    assert point[1] % p != 0
    m = div_mod(3 * point[0] * point[0] + alpha, 2 * point[1], p)
    x = (m * m - 2 * point[0]) % p
    y = (m * (point[0] - x) - point[1]) % p
    return x, y


def ec_mult(m: int, point: ECPoint, alpha: int, p: int) -> ECPoint:
    """
    Multiplies by m a point on the elliptic curve with equation y^2 = x^3 + alpha*x + beta mod p.
    Assumes the point is given in affine form (x, y) and that 0 < m < order(point).
    """
    if m == 1:
        return point
    if m % 2 == 0:
        return ec_mult(m // 2, ec_double(point, alpha, p), alpha, p)
    return ec_add(ec_mult(m - 1, point, alpha, p), point, p)


def get_random_private_key() -> int:
    # returns a private key in the range [1, EC_ORDER)
    return secrets.randbelow(EC_ORDER - 1) + 1


def private_key_to_ec_point_on_stark_curve(priv_key: int) -> ECPoint:
    assert 0 < priv_key < EC_ORDER
    return ec_mult(priv_key, EC_GEN, ALPHA, FIELD_PRIME)


def private_to_stark_key(priv_key: int) -> int:
    return private_key_to_ec_point_on_stark_curve(priv_key)[0]