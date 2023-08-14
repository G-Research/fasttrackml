import os
import platform
import re
from setuptools.command.sdist import sdist
from setuptools import setup, Extension, find_packages
from setuptools.command.build_ext import build_ext
from wheel.bdist_wheel import bdist_wheel as _bdist_wheel
import subprocess
import tarfile

with open("README.md", "r") as f:
    readme = f.read()

def get_version():
    path = './dist/fasttrackml_linux_x86_64' if platform.system() == "Linux" else './dist/fasttrackml_macos_x86_64'
    bin = tarfile.open(f'{path}.tar.gz')
    bin.extractall(path)
    bin.close()
    version_string = subprocess.check_output(
        [f"{path}/bin", "--version"], universal_newlines=True
    ).strip()
    matches = re.findall(r'\d+\.\d+\.\d+', version_string)
    return matches[0]

def get_fml_executable():
     for file_name in os.listdir("bin"):
         if "fml" in file_name:
             return os.path.join("bin", file_name)

class FmlExtension(Extension):
    """Extension for `fml`"""

class FmlBuildExt(build_ext):
    def build_extension(self, ext: Extension) -> None:
        if not isinstance(ext, FmlExtension):
            return super().build_extension(ext)

class bdist_wheel(_bdist_wheel):

        def finalize_options(self):
            _bdist_wheel.finalize_options(self)
            # Mark us as not a pure python package
            self.root_is_pure = False

        def get_tag(self):
            python, abi, plat = _bdist_wheel.get_tag(self)
            # We don't contain any python source
            python, abi = 'py3', 'none'
            return python, abi, plat

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
    ext_modules=[
        FmlExtension(name="fml", sources=["cmd/*"]),
    ],
    cmdclass=dict(
        bdist_wheel=bdist_wheel,
        build_ext=FmlBuildExt,
    ),
)
