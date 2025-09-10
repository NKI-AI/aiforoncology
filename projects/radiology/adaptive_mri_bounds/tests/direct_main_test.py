"""
Test for the main.py script that verifies it runs without crashing.

This test directly imports the main function and tests it in the same process.
The difference between this test and main_test.py:
- This test was always intended to directly import and test the main function
- main_test.py was originally designed to test the script by running it as a subprocess
  through Bazel, but now both tests use the direct import approach
- Both tests are kept as separate examples of testing approaches, though they currently
  perform the same function
"""

import tempfile
import sys
import io

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


# Import the main function from the main module
# In a Bazel context, the import path is relative to the workspace
try:
    from projects.radiology.adaptive_mri_bounds.main import main
except ImportError:
    # Define a stub for linting purposes
    def main(save_dir):
        """Stub for linting purposes."""
        pass


def test_main_function():
    """Test that the main function can be executed successfully with a temporary directory."""
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
