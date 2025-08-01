extern crate cbindgen;

use std::env;

fn main() {
    let crate_dir = env::var("CARGO_MANIFEST_DIR").unwrap();

    cbindgen::Builder::new()
        .with_crate(crate_dir)
        .with_language(cbindgen::Language::C)
        .generate()
        .map_or_else(
            |error| match error {
                cbindgen::Error::ParseSyntaxError { .. } => {}
                e => panic!("{:?}", e),
            },
            |bindings| {
                bindings.write_to_file("target/include/fuel_ffi.h");
            },
        );
}
