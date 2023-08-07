import os
import platform
from setuptools.command.sdist import sdist
from setuptools import setup, Extension, find_packages
from setuptools.command.build_ext import build_ext
from wheel.bdist_wheel import bdist_wheel
import subprocess
import logging

with open("README.md", "r", encoding="utf-8") as f:
    readme = f.read()

def get_fml_executable():
    system = platform.system()
    return "bin/fml.exe" if system == "Windows" else "bin/fml"

class FmlExtension(Extension):
    """Extension for `fml`"""

class FmlBuildExt(build_ext):
    def build_extension(self, ext: Extension) -> None:
        if not isinstance(ext, FmlExtension):
            return super().build_extension(ext)
def get_version():
    # Remove prefix v in versioning
    version = subprocess.check_output(
        [f'./bin/{get_fml_executable()}', "version", "--short"], universal_newlines=True
    ).strip()
    ver = version.rsplit(" ", 1)[-1][1:]
    return ver

setup(
    name="fml",
    version="1.0.0",
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
         build_ext=FmlBuildExt,
         sdist=sdist,
         bdist_wheel=bdist_wheel,
     ),
)
