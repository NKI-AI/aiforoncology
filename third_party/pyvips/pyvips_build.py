import sys
import os

# Add the directory containing vdecls.py to the Python path
current_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.append(current_dir)

import vdecls
from cffi import FFI


def main():
    # Parse arguments for output path
    if len(sys.argv) < 2:
        print("Usage: python3 script.py <output_path>")
        sys.exit(1)

    output_path = sys.argv[1]

    # Version information
    major = 8
    minor = 16
    micro = 0

    # Initialize FFI
    ffibuilder = FFI()
    ffibuilder.set_source(
        "_libvips",
        r"""
        #include <vips/vips.h>
        """,
    )

    # Features and configurations
    features = {
        "major": major,
        "minor": minor,
        "micro": micro,
        "api": True,
    }

    # Add C declarations
    ffibuilder.cdef(vdecls.cdefs(features))

    # Emit the C file at the specified path
    print(f"Generating C code at {output_path}...")
    ffibuilder.emit_c_code(output_path)
    print(f"Generated: {output_path}")


if __name__ == "__main__":
    main()
