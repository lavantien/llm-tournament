# LLM Tournament

A simple and blazingly-fast real-time web app to manage prompts and conduct tournaments for LLMs. Sleek UI/UX with zero bloat.

## Overview

This application simplifies the evaluation of Large Language Models (LLMs) with a user-friendly interface and real-time capabilities. It allows for comprehensive prompt management, model evaluation, and result tracking, all within a responsive and intuitive design.

### Key Features

-   **Real-time Updates**: Leverages WebSockets to provide instant updates on the results page, ensuring users have the latest data at their fingertips.
-   **Dynamic UI**: The user interface is crafted to be both responsive and intuitive, enhancing user experience.
-   **Prompt Management**:
    -   **Add, Edit, Delete**: Full control over prompt creation and management.
    -   **Multiline Input**: Supports multiline input for detailed and complex prompts.
    -   **Markdown Rendering**: Renders prompts in Markdown, allowing for rich text formatting.
    -   **Reorder Prompts**: Drag and drop functionality to easily reorder prompts.
-   **Model Evaluation**:
    -   **Pass/Fail Tracking**: Efficiently tracks pass/fail results for each model against each prompt.
    -   **Total Scores and Pass Percentages**: Displays comprehensive performance metrics for each model.
-   **Data Persistence**: Utilizes JSON files for robust storage of prompts and results.
-   **Import/Export**:
    -   **Prompts and Results**: Supports importing and exporting of prompts and results in CSV format for easy data management.
-   **Filtering**:
    -   **Model Filtering**: Allows filtering of results by model to streamline analysis.
-   **Model Management**:
    -   **Add, Edit, Delete**: Full control over model creation and management.
-   **Result Management**:
    -   **Reset Results**: Allows users to reset all results.
    -   **Refresh Results**: Allows users to refresh all results.

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
