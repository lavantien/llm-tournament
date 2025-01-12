# LLM Tournament Design Document

## **Overview**

The **LLM Tournament** is a real-time web application designed to manage prompts, evaluate Large Language Models (LLMs), and track results in a user-friendly and responsive interface. The application leverages WebSockets for real-time updates, JSON for data persistence, and a lightweight tech stack (Go, HTML, CSS, JS) for performance and simplicity.

---

## **Key Features**

1. **Real-Time Updates**: WebSocket integration ensures instant updates on the results page.
2. **Prompt Management**:
   - Add, edit, delete, and reorder prompts.
   - Multiline input and Markdown rendering for rich text formatting.
3. **Model Evaluation**:
   - Pass/fail tracking for each model against each prompt.
   - Total scores and pass percentages for performance metrics.
4. **Data Persistence**: JSON files store prompts, results, and configurations.
5. **Import/Export**: CSV support for prompts and results.
6. **Filtering**: Filter results by model for streamlined analysis.
7. **Model Management**: Add, edit, and delete models.
8. **Result Management**: Reset and refresh results.
9. **UI/UX**: Sleek, responsive, and intuitive design with zero bloat.

---

## **Architecture**

The application follows a **modular architecture** with clear separation of concerns. The backend is written in Go, while the frontend uses HTML, CSS, and JavaScript. WebSockets enable real-time communication between the client and server.

### **Tech Stack**

- **Backend**: Go (HTTP server, WebSocket server, JSON handling).
- **Frontend**: HTML, CSS, JavaScript (dynamic UI, Markdown rendering, drag-and-drop functionality).
- **Data Storage**: JSON files (`prompts.json`, `results.json`, etc.).
- **Real-Time Communication**: WebSockets.
- **Development Tools**: Aider (with Mistral Large or Gemini 2.0 Flash API).

---

## **Components**

The application is divided into the following logical components:

### **1. Backend**

- **Server**: Handles HTTP requests, WebSocket connections, and JSON file operations.
- **API Endpoints**:
  - `/prompts`: CRUD operations for prompts.
  - `/models`: CRUD operations for models.
  - `/results`: CRUD operations for results.
  - `/ws`: WebSocket endpoint for real-time updates.
- **JSON File Management**:
  - `prompts.json`: Stores prompts and their metadata.
  - `results.json`: Stores evaluation results.
  - `models.json`: Stores model configurations.

### **2. Frontend**

- **Pages**:
  - **Prompt Manager**: Add, edit, delete, and reorder prompts.
  - **Results Page**: Display pass/fail results, total scores, and pass percentages.
  - **Profiles Page**: Manage system prompts and link them to prompts.
  - **Prompt Suites**: Manage and switch between different prompt suites.
  - **Model Suites**: Manage and switch between different model suites.
- **UI Components**:
  - **Prompt List**: Displays prompts with Markdown rendering and drag-and-drop reordering.
  - **Result Table**: Displays evaluation results with filtering options.
  - **Detail Prompt Page**: Shows prompt details, solution, and pass/fail toggles (in detailed mode).
- **Dynamic Features**:
  - Real-time updates via WebSockets.
  - Multiline input and Markdown rendering for prompts.
  - Filtering and sorting options for results.

### **3. Data Flow**

1. **User Interaction**:
   - User interacts with the UI (e.g., adds a prompt, evaluates a model).
2. **API Requests**:
   - Frontend sends HTTP requests to the backend for CRUD operations.
3. **WebSocket Updates**:
   - Backend broadcasts updates to all connected clients via WebSocket.
4. **Data Persistence**:
   - Backend reads/writes data to JSON files.
5. **Frontend Rendering**:
   - Frontend updates the UI dynamically based on API responses and WebSocket messages.

---

## **Workflows**

### **1. Prompt Management**

- **Add Prompt**:

  1. User clicks "Add Prompt" button.
  2. Frontend displays a modal with multiline input and Markdown preview.
  3. User submits the prompt.
  4. Frontend sends a POST request to `/prompts`.
  5. Backend saves the prompt to `prompts.json`.
  6. Backend broadcasts the update via WebSocket.
  7. Frontend updates the prompt list.

- **Edit/Delete Prompt**:

  - Similar workflow with PUT/DELETE requests to `/prompts`.

- **Reorder Prompts**:
  1. User drags and drops a prompt.
  2. Frontend sends a PUT request to `/prompts/reorder`.
  3. Backend updates the order in `prompts.json`.
  4. Backend broadcasts the update via WebSocket.
  5. Frontend updates the prompt list.

### **2. Model Evaluation**

- **Evaluate Model**:
  1. User clicks a cell in the result table.
  2. Frontend toggles pass/fail status.
  3. Frontend sends a PUT request to `/results`.
  4. Backend updates `results.json`.
  5. Backend broadcasts the update via WebSocket.
  6. Frontend updates the result table.

### **3. Import/Export**

- **Import Prompts**:

  1. User uploads a CSV file.
  2. Frontend sends a POST request to `/prompts/import`.
  3. Backend parses the CSV and updates `prompts.json`.
  4. Backend broadcasts the update via WebSocket.
  5. Frontend updates the prompt list.

- **Export Results**:
  1. User clicks "Export Results" button.
  2. Frontend sends a GET request to `/results/export`.
  3. Backend generates a CSV file and sends it to the frontend.
  4. Frontend triggers a download.

---

## **TODO/Roadmap Integration**

### **1. Refactor**

- **Prompt Suite Refactor**:
  - Refactor the prompt suite to be more robust and streamlined, with a focus on modularity.
  - Add a **handbook composition prompt** for better documentation.
- **Codebase Modularity**:
  - Split the codebase into smaller, reusable modules (e.g., `prompts.go`, `models.go`, `results.go`).

### **2. Search Prompt**

- **Functionality**:
  - Add a search bar in the prompt manager to filter prompts by keyword.
  - Located in the rightmost column of the filter row.
- **Implementation**:
  - Add a new API endpoint `/prompts/search` for searching prompts.
  - Update the frontend to include a search input and handle filtering.

### **3. Prompt's Solution**

- **Functionality**:
  - Add a solution field to each prompt.
  - Render the solution alongside the prompt in a 3:1 ratio (prompt region: solution region).
- **Implementation**:
  - Modify `prompts.json` to include a `solution` field.
  - Update the prompt manager UI to include a solution input and rendering area.

### **4. Result-Prompt Integration**

- **Detailed Mode**:
  - Add a button to toggle between **Simple Mode** and **Detailed Mode**.
  - In **Detailed Mode**, clicking a cell in the result table opens a **Detail Prompt Page**.
  - The **Detail Prompt Page** displays the prompt, solution, and a pass/fail toggle.
- **Implementation**:
  - Add a new endpoint `/results/detail` for fetching detailed results.
  - Create a new frontend page `detail-prompt.html` for the detailed view.

### **5. Profiles Page**

- **Functionality**:
  - Add a **Profiles Page** to manage system prompts.
  - Link prompts to profiles via a dropdown selection in the prompt manager.
  - Display the profile name next to each prompt (e.g., `1. (Reasoning) Prompt Content`).
- **Implementation**:
  - Add a new JSON file `profiles.json` to store profiles.
  - Create a new frontend page `profiles.html` for profile management.

### **6. Prompt Suites**

- **Functionality**:
  - Replace the page title with a group of buttons (New, Edit, Delete) and a dropdown for selecting prompt suites.
  - Each suite is stored in a separate JSON file (e.g., `prompts-default.json`, `prompts-suite1.json`).
- **Implementation**:
  - Add new endpoints `/prompts/suites` for managing suites.
  - Update the frontend to include suite management UI.

### **7. Model Suites**

- **Functionality**:
  - Similar to prompt suites, but for models.
  - Allow switching between model suites to render different result tables.
- **Implementation**:
  - Add new endpoints `/models/suites` for managing suites.
  - Update the frontend to include model suite management UI.

---

## **Modular File Structure**

The codebase is organized into the following files:

### **Backend**

1. `main.go`: Entry point for the application.
2. `server.go`: Handles HTTP and WebSocket servers.
3. `prompts.go`: Manages prompt-related operations.
4. `models.go`: Manages model-related operations.
5. `results.go`: Manages result-related operations.
6. `websocket.go`: Handles WebSocket connections and broadcasts.
7. `profiles.go`: Manages profile-related operations (future).
8. `suites.go`: Manages prompt and model suites (future).

### **Frontend**

1. `index.html`: Main page with navigation links.
2. `prompt-manager.html`: Prompt management page.
3. `results.html`: Results page.
4. `profiles.html`: Profiles page (future).
5. `detail-prompt.html`: Detail prompt page (future).
6. `styles.css`: Global styles.
7. `script.js`: JavaScript for dynamic UI and API interactions.

### **Data**

1. `prompts.json`: Stores prompts.
2. `results.json`: Stores results.
3. `models.json`: Stores models.
4. `profiles.json`: Stores profiles (future).
5. `prompts-<suite-name>.json`: Stores prompt suites (future).
6. `models-<suite-name>.json`: Stores model suites (future).
