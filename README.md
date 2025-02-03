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

## 🔑 Key Features

### 🚀 Core Functionality

- **Real-time Evaluation Engine**: WebSocket-powered instant updates for results and metrics.
- **Modular Test Suites**: Independent prompt and model configurations for different scenarios.
- **Comprehensive Data Management**: JSON-based storage with CSV import/export capabilities.
- **Multi-stage Evaluation**: Supports various scoring schemes, including binary (pass/fail) and gradual (0-100 scale) evaluations.
- **Dynamic Result Visualization**: Gradual color changes in the results table to reflect score ranges.
- **Model Search**: Ability to search for specific models in the results table.
- **Advanced Statistics**: Comprehensive performance analytics with interactive charts and tier rankings:
  - Tiered ranking system (Transcendent to Beginner)
  - Stacked score breakdown visualization (20/50/100-point scores)
  - Historical performance comparisons
  - Interactive Chart.js visualizations
- **Tier System**: Auto-classification of models into skill tiers based on aggregate scores
- **Score Analytics**: Detailed breakdown of scoring patterns across different evaluation criteria

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

## 🎓 Tutorial

### 🚀 Getting Started

1. **Setup**: Run `make run` then visit <http://localhost:8080>
2. **Core Pages** (accessible via top navigation):
   - 📝 **Prompts**: Manage evaluation questions and scoring criteria
   - 📊 **Results**: View model performance scores and comparisons
   - 📈 **Stats**: Analyze aggregate metrics and tier rankings
   - 👤 **Profiles**: Configure evaluation personas/scoring profiles

### 🔄 Common Workflows

#### 🧪 Running Evaluations

1. Create prompts in _Prompts_ page (click "+ New Prompt")
2. Configure AI models in _Results_ page (click "+ New Model")
3. Score responses in _Evaluate_ page (accessible from Results table)
4. Track live updates in _Stats_ page during evaluation

#### 📦 Managing Content

- **Bulk Operations**:
  - Check multiple items -> Click "Bulk Actions"
  - Drag to reorder prompts (in Prompts page)
- **Import/Export**:
  - Use CSV buttons in page headers
  - Preserves scoring history and metadata

### 💡 UI/UX Design Philosophy

- **Visual Hierarchy**:
  - Critical actions (Delete, Evaluate) in red
  - Primary actions (Add New) in green
  - Gradual color transitions in scores for quick assessment
- **Realtime Updates**:
  - Auto-refreshing results tables
  - WebSocket-powered score updates
  - Collaborative editing indicators
- **Mobile Optimization**:
  - Collapsible action menus on small screens
  - Touch-friendly drag handles for reordering
  - Large tap targets for scoring buttons

### 📝 Page Breakdown

#### **Prompts Page**

- 🔍 **Search/Filter**: Top-right search bar with profile filters
- ➕ **Add New**: Supports markdown formatting and solution references
- ↔️ **Reorder**: Drag handle (≡) on left of each prompt
- 📁 **Suites**: Manage prompt sets via "Suites" dropdown

#### **Results Page**

- 🏷️ **Model Cards**: Click any model name to edit metadata
- 🔄 **Evaluate**: Orange button launches scoring interface
- 📉 **Trend Lines**: Hover over scores to see historical changes
- 📤 **Export**: Download CSV with all scoring history

#### **Stats Page**

- 🥇 **Tier System**: Automatic classification based on total scores
- 📊 **Score Breakdown**: Interactive pie/bar charts (click to filter)
- 🏅 **Advanced Metrics**: Hover over chart elements for detailed stats

#### **Profiles Page**

- 🎚️ **Preset Management**: Create scoring profiles for different eval scenarios
- 📚 **Profile Attribution**: Assign prompts to specific profiles
- 💡 **Template System**: Clone existing profiles for quick setup

## 🛠️ Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with
  - free/unlimited APIs: Gemini 2.0, Gemini 2.0 Flash, Gemini 2.0 Flash Thinking, Codestral 2501, Mistral Large Latest.
  - paid APIs: DeepSeek V3 since v1.1, DeepSeek R1

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

- Prompts import.
- Rename prompt suite.

### 🔧 Non-Functional

- Make another prompt suite for vision LLMs.

### 🔧 Functional

- Incorporate Promptmetheus's core features.
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
