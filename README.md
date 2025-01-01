# LLM Tournament

A simple and blazingly-fast real-time web app to manage prompts and conduct tournaments for LLMs. Sleek UI/UX with zero bloat.

## Purpose

This application is designed to facilitate the evaluation of Large Language Models (LLMs) by allowing users to:

- Create and manage a list of prompts.
- Evaluate LLMs against these prompts.
- Track the performance of each LLM.
- Export and import prompts and results.
- Reorder prompts using drag and drop.

## Features

- **Real-time Updates**: Uses WebSockets for instant updates on the results page.
- **Dynamic UI**: The user interface is designed to be responsive and intuitive.
- **Prompt Management**: Allows users to add, edit, delete, and reorder prompts.
- **Model Evaluation**: Provides a simple way to track pass/fail results for each model against each prompt.
- **Result Tracking**: Displays total scores and pass percentages for each model.
- **Data Persistence**: Uses JSON files to store prompts and results.
- **Import/Export**: Supports importing and exporting prompts and results as CSV files.
- **Filtering**: Allows filtering of results by model.
- **Drag and Drop**: Enables reordering of prompts using drag and drop functionality.

## Stack

- **Tech**: Go, WebSockets, Built-in Template, HTML, CSS, JS, and database in JSON.
- **Assistant**: Aider with Gemini 2.0 Flash Exp free API.

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
