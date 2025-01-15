# 🏆 LLM Tournament

![Banner](./assets/banner.png)

A powerful, real-time web application designed to manage prompts, evaluate Large Language Models (LLMs), and track results efficiently. Experience a seamless and intuitive user interface with real-time updates, ensuring a smooth and productive evaluation process.

## 📚 Table of Contents

- [Overview](#-overview)
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

A powerful, real-time web application designed to manage prompts, evaluate Large Language Models (LLMs), and track results efficiently. Experience a seamless and intuitive user interface with real-time updates, ensuring a smooth and productive evaluation process. The **LLM Tournament** is a comprehensive, real-time web application designed to streamline the evaluation of Large Language Models (LLMs). It offers robust prompt management, efficient model evaluation, and detailed result tracking, all within a responsive and intuitive interface. The application leverages WebSockets for instant updates, ensuring users have the latest data at their fingertips. With features like real-time updates, dynamic UI, and data persistence, the **LLM Tournament** simplifies the evaluation process, making it easier for users to manage prompts, evaluate models, and track results efficiently.

## 🔑 Key Features

- **🔄 Real-time Updates**: Leverages WebSockets to provide instant updates on the results page, ensuring users have the latest data at their fingertips.
- **🖥️ Dynamic UI**: The user interface is crafted to be both responsive and intuitive, enhancing user experience.
- **📝 Prompt Management**:
  - **➕ Add, ✏️ Edit, ❌ Delete, 🔄 Move**: Full control over prompt creation and management.
  - **🔍 Prompt Solution**: Freely manage prompt's content and solution.
  - **📄 Multiline Input**: Supports multiline input for detailed and complex prompts.
  - **📝 Markdown Rendering**: Renders prompts in Markdown, allowing for rich text formatting.
  - **🔄 Reorder Prompts**: Drag and drop functionality to easily reorder prompts.
  - **🔍 Search Prompts**: Full text search prompt content.
  - **🗑️ Bulk Delete Prompts**: Allows users to delete multiple prompts at once.
- **📊 Model Evaluation**:
  - **✅ Pass/Fail Tracking**: Efficiently tracks pass/fail results for each model against each prompt.
  - **🏆 Total Scores and Pass Percentages**: Displays comprehensive performance metrics for each model.
- **💾 Data Persistence**: Utilizes JSON files for robust storage of prompts and results.
- **📥 Import/Export**:
  - **📝 Prompts and 📊 Results**: Supports importing and exporting of prompts and results in CSV format for easy data management.
- **🔍 Filtering**:
  - **🏆 Model Filtering**: Allows filtering of results by model to streamline analysis.
- **🏆 Model Management**:
  - **➕ Add, ✏️ Edit, ❌ Delete**: Full control over model creation and management.
- **📊 Result Management**:
  - **🔄 Reset Results**: Allows users to reset all results.
  - **🔄 Refresh Results**: Allows users to refresh all results.
  - **📥 Export Results**: Allows users to export results in CSV format.
  - **📥 Import Results**: Allows users to import results from a CSV file.
- **📝 Prompt Suites**:
  - **➕ Create, ✏️ Edit, ❌ Delete, 🔄 Select**: Full control over prompt suite creation, editing, deletion, and selection.
- **📝 Profiles**:
  - **➕ Add, ✏️ Edit, ❌ Delete**: Full control over profile creation and management.
  - **🔄 Reset Profiles**: Allows users to reset all profiles.
  - **🔍 Search Profiles**: Full text search profile content.

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
