def test_basic_math():
    """Sanity check: math should work as expected."""
    assert 2 + 2 == 4
    assert (10 / 5) == 2
    assert sum([1, 2, 3]) == 6
    assert pow(3, 2) == 9


def test_string_behavior():
    """Basic built-in behavior."""
    s = "hip3_pusher"
    assert s.startswith("hip3")
    assert "_" in s
    assert s.replace("_", "-") == "hip3-pusher"
