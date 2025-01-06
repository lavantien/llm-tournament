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
  - **Multiline Input**: Supports multiline input for prompts.
  - **Markdown Rendering**: Supports Markdown rendering for prompts.
- **Model Evaluation**: Provides a simple way to track pass/fail results for each model against each prompt.
- **Result Tracking**: Displays total scores and pass percentages for each model.
- **Data Persistence**: Uses JSON files to store prompts and results.
- **Import/Export**: Supports importing and exporting prompts and results as CSV files.
- **Filtering**: Allows filtering of results by model.
- **Drag and Drop**: Enables reordering of prompts using drag and drop functionality.

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

Here are some potential features that could be added to the application:

**Usability and User Experience:**

- **Search Functionality**: Implement a search bar on the prompt list page to quickly find prompts based on keywords. Allow searching for models on the results page.
- **Prompt Versioning**: Implement a system to track changes to prompts over time, allowing users to revert to previous versions if needed.
- **Model Comparison**: Provide a view that directly compares the performance of two or more models side-by-side.
- **Advanced Statistics**: Calculate and display more detailed statistics, such as the average score per prompt, standard deviation of scores, etc.

**Integration and Extensibility:**

- **Integration with LLM Providers**: Integrate directly with popular LLM providers (e.g., OpenAI, Google AI) to allow users to run tests directly from the application.
- **Customizable Result Metrics**: Allow users to define their own metrics for evaluating LLM performance beyond simple pass/fail.

**Other Considerations:**

- **Accessibility**: Ensure the application is accessible to users with disabilities by following accessibility guidelines (e.g., WCAG).
- **Internationalization**: Support multiple languages to make the application usable by a wider audience.
- **Performance Optimization**: Continuously monitor and optimize the application's performance, especially as the amount of data grows.
