# 🏆 LLM Tournament

LLM Tournament is a streamlined, real-time evaluation platform for Large Language Models. It offers modular test suites, powerful prompt management, and detailed analytics—all built for performance, scalability, and ease of use.

Key technical highlights:

- Lightweight and blazingly-fast due to pure Go Template without any bloat, single binary
- Real-time evaluation engine powered by WebSocket
- Horizontally scalable architecture with stateless components
- Efficient data persistence layer with JSON-based storage
- Responsive frontend built on modern web standards

## 📚 Table of Contents

- 🔑 [Key Features](#-key-features)
- 🎓 [Tutorial](#-tutorial)
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

## Key Features

- **Real-time Evaluation:** Instant updates via WebSocket for immediate feedback.
- **Modular Test Suites:** Easily manage evaluation prompts, models, and profiles.
- **Advanced Analytics:** Interactive charts, tiered rankings, and detailed score breakdowns.
- **Efficient Data Management:** Robust JSON storage coupled with seamless CSV import/export.
- **Intuitive Workflow:** Bulk operations, drag-and-drop prompt reordering, and collaborative functionality.

## Getting Started

1. **Run the Application**
   - Execute `make run` or run `./release/llm-tournament`
   - Open your browser at [http://localhost:8080](http://localhost:8080)

## Development

- Ensure you have Go (and Python for tooling) installed.
- Duplicate `./.aider.conf.yml.example` to `./.aider.conf.yml` and add your API key.
- Use `make aiderupdate` to update dependencies.

## 🛠️ Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with
  - free/unlimited APIs: Gemini 2.0 Flash, Codestral 2501, Mistral Large Latest.
  - paid APIs: DeepSeek V3 since v1.1, DeepSeek R1, o3-mini (high), o1 (high), Claude 3.7 Sonnet.

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

### 🔧 Non-Functional

- Make another prompt suite for vision LLMs.

### 🔧 Functional

- Add RAG and Web search agentic system under `./tools/ragweb_agent/`.
- Update the features section about the tools.

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
