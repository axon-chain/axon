from setuptools import setup, find_packages

setup(
    name="axon-sdk",
    version="0.3.0",
    description="Python SDK for the Axon AI Agent blockchain",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    author="Axon Chain",
    url="https://github.com/axon-chain/axon",
    packages=find_packages(),
    python_requires=">=3.8",
    install_requires=[
        "web3>=6.0.0",
        "eth-account>=0.9.0",
    ],
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Programming Language :: Python :: 3",
    ],
)
