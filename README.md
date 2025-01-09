# LLM Tournament

A simple and blazingly-fast real-time web app to manage prompts and conduct tournaments for LLMs. Sleek UI/UX with zero bloat.

## Overview

This application simplifies the evaluation of Large Language Models (LLMs) with a user-friendly interface and real-time capabilities. It allows for comprehensive prompt management, model evaluation, and result tracking, all within a responsive and intuitive design.

### Key Features

- **Real-time Updates**: Leverages WebSockets to provide instant updates on the results page, ensuring users have the latest data at their fingertips.
- **Dynamic UI**: The user interface is crafted to be both responsive and intuitive, enhancing user experience.
- **Prompt Management**:
  - **Add, Edit, Delete**: Full control over prompt creation and management.
  - **Multiline Input**: Supports multiline input for detailed and complex prompts.
  - **Markdown Rendering**: Renders prompts in Markdown, allowing for rich text formatting.
  - **Reorder Prompts**: Drag and drop functionality to easily reorder prompts.
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

## Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with Mistral Large Latest or Gemini 2.0 Flash Exp free API.

## UI

![prompt-manager-page](./assets/ui-prompt-manager.png)

![result-page](./assets/ui-result-page.png)

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

## Future Enhancements

- Refacxtor the codebase to be modular
- Prompt input with solution.
- Profiles page similar to Prompts to store system prompts.
- Link prompts with profiles, prompt input and edit have drop down selection for profile.
- Display profile of each prompt after the numbering, e.g. 1. (Reasoning) prompt content
