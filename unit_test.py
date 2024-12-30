import subprocess
import sys
import re

# ANSI escape codes for colors
class Colors:
    RESET = "\033[0m"
    BLUE = "\033[94m"
    GREEN = "\033[92m"
    RED = "\033[91m"
    CYAN = "\033[96m"
    BOLD = "\033[1m"

def get_packages_unit_test():
    """
    Get the list of Go packages excluding specified directories.
    """
    EXCLUDE_DIRS = ["cmd", "cli", "migration", "vendor"]

    print(f"{Colors.BLUE}{Colors.BOLD}Collecting packages...{Colors.RESET}")
    try:
        result = subprocess.run(
            ["go", "list", "./..."],
            capture_output=True,
            text=True,
            check=True
        )
        packages = result.stdout.splitlines()
        # Filter out excluded directories
        filtered_packages = [
            pkg for pkg in packages if not any(exclude in pkg for exclude in EXCLUDE_DIRS)
        ]
        print(f"{Colors.GREEN}Found {len(filtered_packages)} packages to test.{Colors.RESET}")
        return filtered_packages
    except subprocess.CalledProcessError as e:
        print(f"{Colors.RED}Error listing packages: {e.stderr}{Colors.RESET}")
        sys.exit(1)

def is_cloud_provider_package(pkg):
    """
    Check if the package matches the pattern for `pkg/provider/{CLOUD_PROVIDER}`.
    """
    match = re.match(r".*/pkg/provider/([^/]+)$", pkg)
    return match.group(1) if match else None

def run_tests(packages):
    """
    Run `go test -v` on each package. If the package matches `pkg/provider/{CLOUD_PROVIDER}`,
    include the appropriate compiler tag.
    """
    for pkg in packages:
        # Check if the package matches `pkg/provider/{CLOUD_PROVIDER}`
        cloud_provider = is_cloud_provider_package(pkg)
        if cloud_provider:
            go_tag = f"testing_{cloud_provider.lower()}"
            print(f"\n{Colors.CYAN}Running tests for package: {pkg} with tag: {go_tag}{Colors.RESET}")
            command = ["go", "test", "-v", "-tags", go_tag, pkg]
        else:
            print(f"\n{Colors.CYAN}Running tests for package: {pkg}{Colors.RESET}")
            command = ["go", "test", "-v", pkg]

        try:
            result = subprocess.run(command, check=True)
            if result.returncode == 0:
                print(f"{Colors.GREEN}✔ Tests passed for package: {pkg}{Colors.RESET}")
        except subprocess.CalledProcessError:
            print(f"{Colors.RED}✘ Tests failed for package: {pkg}{Colors.RESET}")
            return False
    return True

if __name__ == "__main__":
    # Step 1: Get the packages
    packages = get_packages_unit_test()
    if not packages:
        print(f"{Colors.RED}No packages to test.{Colors.RESET}")
        sys.exit(1)

    # Step 2: Run tests for each package
    all_tests_passed = run_tests(packages)

    # Step 3: Final message
    if all_tests_passed:
        print(f"\n{Colors.GREEN}{Colors.BOLD}🎉 All tests passed successfully!{Colors.RESET}")
    else:
        print(f"\n{Colors.RED}{Colors.BOLD}🚨 Some tests failed. Check the output above for details.{Colors.RESET}")
        sys.exit(1)
