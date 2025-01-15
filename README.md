# 🏆 LLM Tournament

![Banner](./assets/banner.png)

The **LLM Tournament** is a powerful, real-time web application designed to streamline the evaluation of Large Language Models (LLMs). It offers robust prompt management, efficient model evaluation, and detailed result tracking, all within a responsive and intuitive interface. With features like real-time updates, dynamic UI, data persistence, prompt suites, and profiles, the **LLM Tournament** simplifies the evaluation process, making it easier for users to manage prompts, evaluate models, and track results efficiently.

## 📚 Table of Contents

- [Key Features](#-key-features)
- [Stack](#-stack)
- [UI](#-ui)
- [Run](#-run)
- [Develop](#-develop)
- [Contribute](#-contribute)
- [TODO/Roadmap](#-todoroadmap)
- [Badges](#-badges)
- [Contributors](#-contributors)
- [License](#-license)
- [Contact](#-contact)

## 🔑 Key Features

- **🔄 Real-time Updates**: WebSockets for instant updates on the results page.
- **🖥️ Dynamic UI**: Responsive and intuitive interface.
- **📝 Prompt Management**:
  - **➕ Add, ✏️ Edit, ❌ Delete, 🔄 Move**: Manage prompts.
  - **🔍 Prompt Solution**: Manage prompt content and solution.
  - **📄 Multiline Input**: Detailed and complex prompts.
  - **📝 Markdown Rendering**: Rich text formatting.
  - **🔄 Reorder Prompts**: Drag and drop reordering.
  - **🔍 Search Prompts**: Full text search.
  - **🗑️ Bulk Delete Prompts**: Delete multiple prompts.
- **📊 Model Evaluation**:
  - **✅ Pass/Fail Tracking**: Track pass/fail results.
  - **🏆 Total Scores and Pass Percentages**: Performance metrics.
- **💾 Data Persistence**: JSON files for storage.
- **📥 Import/Export**:
  - **📝 Prompts and 📊 Results**: CSV format for data management.
- **🔍 Filtering**:
  - **🏆 Model Filtering**: Filter results by model.
- **🏆 Model Management**:
  - **➕ Add, ✏️ Edit, ❌ Delete**: Manage models.
- **📊 Result Management**:
  - **🔄 Reset Results**: Reset all results.
  - **🔄 Refresh Results**: Refresh all results.
  - **📥 Export Results**: Export results in CSV format.
  - **📥 Import Results**: Import results from a CSV file.
- **📝 Prompt Suites**:
  - **➕ Create, ✏️ Edit, ❌ Delete, 🔄 Select**: Manage prompt suites.
- **📝 Profiles**:
  - **➕ Add, ✏️ Edit, ❌ Delete**: Manage profiles.
  - **🔄 Reset Profiles**: Reset all profiles.
  - **🔍 Search Profiles**: Full text search.

## 🛠️ Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with free/unlimited APIs: Gemini 2.0 Advanced, Gemini 2.0 Flash, Codestral 2501, Mistral Large Latest.

## 🖼️ UI

![prompt-manager-page](./assets/ui-prompt-manager.png)

![result-page](./assets/ui-result-page.png)

![profile-page](./assets/ui-profile-manager.png)

![prompt-edit-page](./assets/ui-prompt-edit.png)

## 🏃 Run

```bash
make run
```

Then go to <http://localhost:8080>

## 🛠️ Develop

Require Linux environment with Python and Go installed (preferably via Brew).

```bash
make aiderupdate
```

Then tweak `./.aider.conf.yml.example` into `./.aider.conf.yml` with your own API Key.

## 🤝 Contribute

Anyone can just submit a PR and we'll discuss there.

## 📝 TODO/Roadmap

### 🔧 Non-Functional

- Handbook composition prompt.
- Make another prompt suite to test SD 3.5 vs Flux 1 Dev.
- Make another prompt suite for vision LLMs.

## 🏆 Badges

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![GitHub issues](https://img.shields.io/github/issues/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/issues)
[![GitHub stars](https://img.shields.io/github/stars/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/lavantien/llm-tournament)](https://github.com/lavantien/llm-tournament/network)

## 👥 Contributors

<a href="https://github.com/lavantien/llm-tournament/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lavantien/llm-tournament" />
</a>

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact

For any questions or suggestions or collaboration/job inquiries, feel free to reach out to us at [cariyaputta@gmail.com](mailto:cariyaputta@gmail.com).
