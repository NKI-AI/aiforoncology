# AI for Oncology Lab

Welcome to the [**AI for Oncology Lab**](https://aiforoncology.nl/) public repository!  
We are based at [**The Netherlands Cancer Institute**](https://www.nki.nl/), and our mission is to develop **artificial intelligence innovations** for the improvement of **cancer diagnostics and therapy**.

This repository contains all our tools, libraries, and frameworks that make it possible for us to achieve this mission through state-of-the-art AI solutions. We are committed to open science and open source, and we are making our best efforts to make our code and models as accessible as possible.

---

## 🌟 Project Overview

The repository hosts a wide range of projects and libraries related to:

- **Deep learning for oncology**: Models for cancer classification, grading, detection and more.
- **Image processing**: Efficient pipelines for working with gigapixel histopathology and medical images.
- **Dataset management**: Tools for handling large-scale datasets with distributed systems and shared memory.
- **Clinical integration**: Frameworks for translating AI research into clinical practice.

For more details and technical documentation, visit our [**project website**](https://aifo.dev/).

---

## 🚀 Getting Started

### Prerequisites

We use **Bazelisk** to manage Bazel versions for building and testing.  
To install Bazelisk, follow the instructions in the [**Bazelisk GitHub repository**](https://github.com/bazelbuild/bazelisk) or install it using your package manager:

```bash
# Example for MacOS using Homebrew
brew install bazelisk
```

### Build and Test

You can build and test the project with the following commands:

```bash
# Build the entire project
bazelisk build //...

# Run all tests
bazelisk test //...
```

### Platform-Specific Configuration

- **MacOS**: You may need a `.bazelrc.local` configuration file:
  ```bash
  echo "build --config=macos" > .bazelrc.local
  ```
- **Linux**: Adjust your `.bazelrc.local` file as necessary for your distribution.

---

## 👥 Contributors

A huge thank you to our dedicated contributors and collaborators for their invaluable support in alphabetical order:

- [**Joren Brunekreef**](https://github.com/JorenB)
- [**Eric Marcus**](https://github.com/EricMarcus-ai)
- [**Marek Oerlemans**](https://github.com/moerlemans)
- [**Ajey Pai Karkala**](https://github.com/AjeyPaiK)
- [**Jonas Teuwen**](https://github.com/jonasteuwen)
- [**Vivien van Veldhuizen**](https://github.com/VivienvV)
- [**Tim Veenboer**](https://github.com/TimVeenboer)

Interested in contributing? Feel free to explore, fork the repository, and submit pull requests!  
For major contributions, please get in touch with us via our [**lab website**](https://aiforoncology.nl/).

---

## 📚 Learn More

- **Lab website**: [https://aiforoncology.nl/](https://aiforoncology.nl/)
- **Project documentation and details**: [https://aifo.dev/](https://aifo.dev/)

---

## 📜 License

This project is licensed under the [**Apache 2.0 License**](LICENSE). Feel free to use, modify, and distribute it under the terms of the license.

---

Thank you for your interest in our work!  
Together, we are shaping the future of oncology with cutting-edge AI solutions.
