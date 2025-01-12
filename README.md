# LLM Tournament

A simple and blazingly-fast real-time web app to manage prompts and conduct tournaments for LLMs. Sleek UI/UX with zero bloat.

## Overview

This application simplifies the evaluation of Large Language Models (LLMs) with a user-friendly interface and real-time capabilities. It allows for comprehensive prompt management, model evaluation, and result tracking, all within a responsive and intuitive design.

### Key Features

- **Real-time Updates**: Leverages WebSockets to provide instant updates on the results page, ensuring users have the latest data at their fingertips.
- **Dynamic UI**: The user interface is crafted to be both responsive and intuitive, enhancing user experience.
- **Prompt Management**:
  - **Add, Edit, Delete, Move**: Full control over prompt creation and management.
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

## TODO/Roadmap

### Refactor

- Refactor the codebase to be modular.
- Refactor the prompt suite to be more roburst and streamlined, with programming focus.
- Handbook composition prompt.
- Make another prompt suite to test SD 3.5 vs Flux 1 Dev.
- Make another prompt suite for vision LLMs.

### Search Prompt

- Add a search prompt functionality, it will be located in the right most of the filter by order's row.

### Prompt's Solution

- Prompt input with solution; and solution will be rendered along side with the prompt in prompt list (3/4 prompt region, 1/4 solution region).

### Result-Prompt Integration

- Add an button on the left most of the add model row, the button's name has two states: Simple Mode and Detailed Mode, when click will switch between these 2 mode, in simple mode everything stay the same, but in detailed mode the logic is as below.
- In detailed mode, when click on a cell in result table, instead of pass/fail logic, open a detail_prompt page that show the prompt and solution rendered in markdown and a pass/fail toggle and an accept button and a back b utton.

### 3rd Page - Profiles

- Profiles page similar to Prompts to store system prompts.
- Link prompts with profiles, prompt input and edit have drop down selection for profile.
- Display profile of each prompt after the numbering, e.g. 1. (Reasoning) prompt content.

### Prompt Suites

- In place of the page title in the middle (Prompt List), replace it with a group of 3 buttons and a dropdown selection.
- The 2 buttons are: New, Edit, and Delete; the dropdown selection is to choose the current prompt suite to display.
- When click either of the 2 buttons, the appropriate page will be redirected to a similarly action page like others: the new_prompt_suite and edit_prompt_suite pages will have an input field and accept and cancel buttons, the delete_prompt_suite page will have the accept and cancel buttons.
- The starting prompts in `prompts.json` will be the `default` suite.
- Each suite will be stored in a different json: `prompts-<suite-name>.json`

### Model Suites

- The same with prompt suite, but for models and result table rendering.
