import os
from setuptools.command.sdist import sdist
from setuptools import setup, Extension, find_packages
from setuptools.command.build_ext import build_ext
from wheel.bdist_wheel import bdist_wheel


import subprocess
import logging


class FmlExtension(Extension):
    """Extension for `fml`"""


class FmlBuildExt(build_ext):
    def build_extension(self, ext: Extension) -> None:
        if not isinstance(ext, FmlExtension):
            return super().build_extension(ext)

class bdist_wheel_universal(bdist_wheel):
    def get_tag(self):
        *_, plat = super().get_tag()
        return "py2.py3", "none", plat


class SdistCommand(sdist):
    def run(self):
        errno = subprocess.call(["make", "generate-git-tag-info"])
        assert errno == 0, "Failed to generate git tag info"
        if not os.path.isfile(".GIT_TAG_INFO"):
            logging.warning(".GIT_TAG_INFO not found")
        else:
            logging.info(".GIT_TAG_INFO generated")
        super().run()


classifiers = [
    "Development Status :: 3 - Alpha",
    "Topic :: Software Development :: Build Tools",
    "Intended Audience :: Science/Research",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: Apache Software License",
    "Programming Language :: Python :: 3.6",
    "Programming Language :: Python :: 3.7",
    "Programming Language :: Python :: 3.8",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
]


setup(
    name="fml",
    version="1.0.0",
    description="A development environment management tool for data scientists.",
    packages=find_packages(),
    include_package_data=True,
    python_requires=">=3.6",
    data_files=[("bin", ["bin/fml"])],
    classifiers=classifiers,
    zip_safe=False,
    ext_modules=[
        FmlExtension(name="fml", sources=["cmd/*"]),
    ],
    cmdclass=dict(
        build_ext=FmlBuildExt,
        sdist=SdistCommand,
        bdist_wheel=bdist_wheel_universal,
    ),
)