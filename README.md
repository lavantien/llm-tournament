# LLM Tournament

A simple and blazingly-fast real-time web app to manage prompts and conduct tournaments for LLMs. Sleek UI/UX with zero bloat.

## Overview

This application simplifies the evaluation of Large Language Models (LLMs) with a user-friendly interface and real-time capabilities. It allows for comprehensive prompt management, model evaluation, and result tracking, all within a responsive and intuitive design.

### Key Features

- **Real-time Updates**: Leverages WebSockets to provide instant updates on the results page, ensuring users have the latest data at their fingertips.
- **Dynamic UI**: The user interface is crafted to be both responsive and intuitive, enhancing user experience.
- **Prompt Management**:
  - **Add, Edit, Delete, Move**: Full control over prompt creation and management.
  - **Prompt Solution**: Freely manage prompt's content and solution.
  - **Multiline Input**: Supports multiline input for detailed and complex prompts.
  - **Markdown Rendering**: Renders prompts in Markdown, allowing for rich text formatting.
  - **Reorder Prompts**: Drag and drop functionality to easily reorder prompts.
  - **Search Prompts**: Full text search prompt content.
  - **Bulk Delete Prompts**: Allows users to delete multiple prompts at once.
- **Model Evaluation**:
  - **Pass/Fail Tracking**: Efficiently tracks pass/fail results for each model against each prompt.
  - **Total Scores and Pass Percentages**: Displays comprehensive performance metrics for each model.
- **Data Persistence**: Utilizes JSON files for robust storage of prompts and results.
- **Import/Export**:
  - **Prompts and Results**: Supports importing and exporting of prompts and results in CSV format for easy data management.
- **Filtering**:
  - **Model Filtering**: Allows filtering of results by model to streamline analysis.
- **Model Management**:
  - **Add, Edit, Delete**: Full control over model creation and management.
- **Result Management**:
  - **Reset Results**: Allows users to reset all results.
  - **Refresh Results**: Allows users to refresh all results.
  - **Export Results**: Allows users to export results in CSV format.
  - **Import Results**: Allows users to import results from a CSV file.
- **Prompt Suites**:
  - **Create, Edit, Delete, Select**: Full control over prompt suite creation, editing, deletion, and selection.

## Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with Mistral Large Latest or Gemini 2.0 Flash Exp free API.

## UI

![prompt-manager-page](./assets/ui-prompt-manager.png)

![result-page](./assets/ui-result-page.png)

![prompt-edit-page](./assets/ui-prompt-edit.png)

## Run

```bash
make run
```

Then go to <http://localhost:8080>

## Develop

Require Linux environment with Python and Go installed (preferably via Brew).

```bash
make updateaider
```

Then tweak `./.aider.conf.yml.example` into `./.aider.conf.yml` with your own API Key.

## Contribute

Anyone can just submit a PR and we'll discuss there.

## TODO/Roadmap

### Refactor

- Change UI demo pictures after finished profiles page.
- Refactor the prompt suite to be more roburst and streamlined, with programming focus.
- Handbook composition prompt.
- Make another prompt suite to test SD 3.5 vs Flux 1 Dev.
- Make another prompt suite for vision LLMs.

### 3rd Page - Profiles

- Profiles page similar to Prompts to store system prompts.
- Link prompts with profiles, prompt input and edit have drop down selection for profile.
- Display profile of each prompt after the numbering, e.g. 1. (Reasoning) prompt content.

### Result-Prompt Integration

- Add an button on the left most of the add model row, the button's name has two states: Simple Mode and Detailed Mode, when click will switch between these 2 mode, in simple mode everything stay the same, but in detailed mode the logic is as below.
- In detailed mode, when click on a cell in result table, instead of pass/fail logic, open a detail_prompt page that show the prompt and solution rendered in markdown and a pass/fail toggle and an accept button and a back b utton.
