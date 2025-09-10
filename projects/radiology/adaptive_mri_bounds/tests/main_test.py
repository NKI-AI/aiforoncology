"""
Test for the main.py script that verifies it runs without crashing.

This test directly imports the main function and tests it in the same process.
The difference between this test and direct_main_test.py:
- This test was originally intended to run the main script as a subprocess via Bazel command,
  but was changed to directly import the main function due to Bazel-in-Bazel testing limitations.
- Both tests now perform the same function (directly testing the main function) but are kept
  as separate examples of two testing approaches.
"""

import tempfile
import sys
import io
import os

# Handle pytest import (will be available at runtime via Bazel dependencies)
try:
    import pytest
except ImportError:
    # For linting, create a minimal pytest mock
    class pytest:
        @staticmethod
        def fail(msg):
            raise AssertionError(msg)

        @staticmethod
        def main():
            pass


# Add the parent directory to the path to allow imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Import the main function from the main module
# This import will work at runtime, but may show linter errors
try:
    from main import main
except ImportError:
    # Define a stub for linting purposes
    def main(save_dir):
        """Stub for linting purposes."""
        pass


def test_main_execution():
    """Test that the main script can be executed successfully with default parameters."""
    # Create a temporary directory for test outputs
    with tempfile.TemporaryDirectory() as temp_dir:
        try:
            # Capture stdout to check for expected output message
            original_stdout = sys.stdout
            captured_output = io.StringIO()
            sys.stdout = captured_output

            # Call the main function directly
            main(save_dir=temp_dir)

            # Restore stdout
            sys.stdout = original_stdout
            output = captured_output.getvalue()

            # Check that the expected message is in the output
            assert "All figures generated and saved to" in output, "Expected output message not found"
            assert temp_dir in output, f"Expected path {temp_dir} not found in output"

        except Exception as e:
            # Restore stdout if exception occurs
            if sys.stdout != original_stdout:
                sys.stdout = original_stdout
            pytest.fail(f"Main function execution failed with error: {str(e)}")


if __name__ == "__main__":
    pytest.main()
