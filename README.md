# 🏆 LLM Tournament

![Banner](./assets/banner.png)

A high-performance, _blazingly-fast_ evaluation platform for Large Language Models, built with enterprise-grade architecture and real-time capabilities. This platform enables systematic assessment of LLM performance through comprehensive test suites, sophisticated prompt management, and detailed analytics.

## 💡 Overview

LLM Tournament addresses the critical challenge of evaluating and comparing language model performance at scale. Built with a focus on reliability and real-time processing, it provides a robust framework for managing complex evaluation workflows while maintaining high performance and data integrity.

Key technical highlights:

- Lightweight and blazingly-fast due to pure Go Template without any bloat, single binary
- Real-time evaluation engine powered by WebSocket
- Horizontally scalable architecture with stateless components
- Efficient data persistence layer with JSON-based storage
- Responsive frontend built on modern web standards

## 📚 Table of Contents

- 🔑 [Key Features](#-key-features)
- 🛠️ [Stack](#%EF%B8%8F-stack)
- 🖼️ [UI](#%EF%B8%8F-ui)
- 🏃 [Run](#-run)
- 🛠️ [Develop](#%EF%B8%8F-develop)
- 🤝 [Contribute](#-contribute)
- 📝 [TODO/Roadmap](#-todoroadmap)
- 🏆 [Badges](#-badges)
- 👥 [Contributors](#-contributors)
- 📜 [License](#-license)
- 📞 [Contact](#-contact)

## 🔑 Key Features

### 🚀 Core Functionality

- **Real-time Evaluation Engine**: WebSocket-powered instant updates for results and metrics.
- **Modular Test Suites**: Independent prompt and model configurations for different scenarios.
- **Comprehensive Data Management**: JSON-based storage with CSV import/export capabilities.
- **Multi-stage Evaluation**: Supports various scoring schemes, including binary (pass/fail) and gradual (0-100 scale) evaluations.
- **Dynamic Result Visualization**: Gradual color changes in the results table to reflect score ranges.
- **Model Search**: Ability to search for specific models in the results table.

### 📝 Prompt Management

- **Full Lifecycle Control**: Create, edit, delete, and reorder prompts.
- **Rich Content Support**: Markdown formatting and multiline input.
- **Advanced Filtering**: Search by text, filter by profile and order.
- **Bulk Operations**: Delete multiple prompts at once.
- **Solution Tracking**: Attach reference solutions to each prompt.
- **Profile Association**: Tag prompts with evaluation profiles.
- **Improved Prompt Ordering**: Drag and drop interface to reorder prompts.

### 🏆 Model Evaluation

- **Performance Tracking**: Binary (pass/fail) and gradual (0-100 scale) scoring with detailed metrics.
- **Real-time Analytics**: Scores and pass percentages updated instantly.
- **Flexible Filtering**: View results by model or profile.
- **Data Portability**: Import/export results in CSV format.
- **Evaluation Management**: Reset or refresh results as needed.
- **Enhanced Result Actions**: Edit and delete models directly from the results table.

### ⚙️ System Management

- **Prompt Suites**: Create, edit, delete, and switch between different prompt sets.
- **Model Suites**: Manage different model configurations.
- **Profile System**: Define and manage evaluation profiles.
- **Data Integrity**: Automatic backups and version control.
- **Responsive UI**: Modern interface optimized for all devices.
- **Streamlined Navigation**: Quick access to results, prompts, and profiles from the main navigation bar.

### 🔄 Workflow Automation

- **Bulk Operations**: Manage multiple items simultaneously
- **Template System**: Reuse configurations across evaluations
- **Data Migration**: Easy import/export of prompts and results
- **Real-time Sync**: Instant updates across all connected clients

[(Back)](#-table-of-contents)

## 🛠️ Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with
  - free/unlimited APIs: Gemini 2.0 Advanced, Gemini 2.0 Flash, Codestral 2501, Mistral Large Latest.
  - paid deepseek-3-chat API since v1.1

[(Back)](#-table-of-contents)

## 🖼️ UI

![prompt-manager-page](./assets/ui-prompt-manager.png)

![result-page](./assets/ui-result-page.png)

![profile-page](./assets/ui-profile-manager.png)

![prompt-edit-page](./assets/ui-prompt-edit.png)

[(Back)](#-table-of-contents)

## 🏃 Run

```bash
make run
```

```bash
./release/llm-tournament
```

Then go to <http://localhost:8080>

[(Back)](#-table-of-contents)

## 🛠️ Develop

Require Linux environment with Python and Go installed (preferably via Brew).

```bash
make aiderupdate
```

Then tweak `./.aider.conf.yml.example` into `./.aider.conf.yml` with your own API Key.

[(Back)](#-table-of-contents)

## 🤝 Contribute

Anyone can just submit a PR and we'll discuss there.

[(Back)](#-table-of-contents)

## 📝 TODO/Roadmap

### 🔧 Issues

- Result list not auto render after import.
- Switch `,` to `;` so that CSV doesn't cut it out.

### 🔧 Non-Functional

- Make another prompt suite for vision LLMs.

### 🔧 Functional

- Incorporate Promptmetheus's core features.
- Add RAG and Web search agentic system under `./tools/ragweb_agent/`.
- Update the features section about the tools.
- Add a feature to evaluate models using different scoring schemes.
- Add a feature to compare models' performance.

[(Back)](#-table-of-contents)

## 🏆 Badges

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![GitHub issues](https://img.shields.io/github/issues/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/issues)
[![GitHub stars](https://img.shields.io/github/stars/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/network)

[(Back)](#-table-of-contents)

## 👥 Contributors

<a href="https://github.com/lavantien/llm-tournament/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lavantien/llm-tournament" />
</a>

[(Back)](#-table-of-contents)

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

[(Back)](#-table-of-contents)

## 📞 Contact

For any questions or suggestions or collaboration/job inquiries, feel free to reach out to us at [cariyaputta@gmail.com](mailto:cariyaputta@gmail.com).

[(Back)](#-table-of-contents)
