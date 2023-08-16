import os
import subprocess
import re
from setuptools.command.sdist import sdist
from setuptools import setup, Extension, find_packages
from setuptools.command.build_ext import build_ext
from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

with open("README.md", "r", encoding="utf-8") as f:
    readme = f.read()


def get_fml_executable():
    os = subprocess.check_output(["go", "env", "GOOS"]).strip().decode("utf-8")
    return "fml.exe" if os == "windows" else "fml"

def get_version():
    version = os.environ.get("VERSION")
    matches = re.findall(r'\d+\.\d+\.\d+', version)
    return matches[0]

def get_platform():
    os = subprocess.check_output(["go", "env", "GOOS"]).strip().decode("utf-8")
    arch = subprocess.check_output(["go", "env", "GOARCH"]).strip().decode("utf-8")
    print(f"{os}_{arch}")
    if f"{os}_{arch}" == "darwin_amd64":
        return "macosx_10_13_x86_64"
    elif f"{os}_{arch}" == "darwin_arm64":
        return "macosx_11_0_arm64"
    elif f"{os}_{arch}" == "linux_amd64":
        return "manylinux1_x86_64"
    elif f"{os}_{arch}" == "linux_arm64":
        return "manylinux1_aarch64"
    elif f"{os}_{arch}" == "windows_amd64":
        return "win_amd64"

class bdist_wheel(_bdist_wheel):

    def finalize_options(self):
        _bdist_wheel.finalize_options(self)
        # Mark us as not a pure python package
        self.root_is_pure = False

    def get_tag(self):
        python, abi, plat = _bdist_wheel.get_tag(self)
        # We don't contain any python source
        python, abi = 'py3', 'none'
        return 'py3', 'none', get_platform()


setup(
    name="fml",
    version=get_version(),
    description="Rewrite of the MLFlow tracking server with a focus on scalability.",
    long_description=readme,
    packages=find_packages(),
    include_package_data=True,
    data_files=[("bin", [get_fml_executable()])],
    python_requires=">=3.6",
    zip_safe=False,
    ext_modules=[],
    cmdclass=dict(
        sdist=sdist,
        bdist_wheel=bdist_wheel,
    ),
)
