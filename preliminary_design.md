**Step-by-Step Explanation:**

1. **Set Up the Project Structure:**

   - Create a new directory for the project.
   - Inside the directory, create the following files and folders:
     - `main.go`
     - `templates/`
       - `prompt_list.html`
       - `results.html`
     - `data/`
       - `prompts.txt`
       - `results.csv`

2. **Initialize the Web Server:**

   - In `main.go`, import the necessary packages:
     ```go
     import (
         "html/template"
         "net/http"
         "os"
         "strconv"
         "strings"
     )
     ```
   - Define a `Router` function to handle HTTP requests.
   - Start the server on a local port, e.g., `http.ListenAndServe(":8080", nil)`.

3. **Create HTML Templates:**

   - In `prompt_list.html`, design a page to display and manage prompts.
   - Include forms to add, edit, and delete prompts.
   - Use Go template actions to loop through and display prompts.

   - In `results.html`, design a spreadsheet-like table:
     - Rows represent different LLM models.
     - Columns represent pass/fail checkboxes for each prompt.
     - Include a final column for the total score.
     - Use JavaScript for real-time updates on checkbox changes.

4. **Manage Prompts:**

   - Create a function to read prompts from `prompts.txt`:
     ```go
     func readPrompts() []string {
         // Read lines from prompts.txt
     }
     ```
   - Create functions to add, edit, and delete prompts:
     ```go
     func addPrompt(prompt string) {
         // Append prompt to prompts.txt
     }
     func editPrompt(index int, newPrompt string) {
         // Update the prompt at the given index
     }
     func deletePrompt(index int) {
         // Remove the prompt at the given index
     }
     ```

5. **Manage Results:**

   - Create a function to read results from `results.csv`:
     ```go
     func readResults() map[string][]bool {
         // Read CSV and return a map of model names to []bool for pass/fail
     }
     ```
   - Create functions to update results:
     ```go
     func updateResult(model string, promptIndex int, pass bool) {
         // Update the result for the model and prompt
     }
     ```
   - Calculate the total score for each model by summing the passes.

6. **Implement Real-Time Updates:**

   - Use JavaScript to handle checkbox changes on the client side.
   - Send AJAX requests to the server when a checkbox is changed.
   - Update the results file and recalculate scores on the server.
   - Return the updated scores and render them on the page.

7. **Auto-Sort Rows by Total Score:**

   - After calculating the total scores, sort the models in descending order.
   - Re-render the table with the sorted models.

8. **Ensure No External Dependencies:**

   - Use only Go's standard libraries.
   - Avoid third-party packages for routing, templates, etc.

9. **Test the Application:**
   - Run the server and access it through a web browser.
   - Test adding, editing, and deleting prompts.
   - Test marking pass/fail on the results page.
   - Verify that the total scores update correctly and rows are sorted.

**Code Implementation:**

_`main.go`_

```go
package main

import (
    "html/template"
    "net/http"
    "os"
    "strconv"
    "strings"
)

func main() {
    http.HandleFunc("/", router)
    http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
    http.ListenAndServe(":8080", nil)
}

func router(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/prompts" {
        promptListHandler(w, r)
    } else if r.URL.Path == "/results" {
        resultsHandler(w, r)
    } else if r.URL.Path == "/update_result" {
        updateResultHandler(w, r)
    } else {
        http.Redirect(w, r, "/prompts", http.StatusSeeOther)
    }
}

// Read prompts from prompts.txt
func readPrompts() []string {
    data, _ := os.ReadFile("data/prompts.txt")
    prompts := strings.Split(string(data), "\n")
    return prompts
}

// Read results from results.csv
func readResults() map[string][]bool {
    data, _ := os.ReadFile("data/results.csv")
    lines := strings.Split(string(data), "\n")
    results := make(map[string][]bool)
    prompts := readPrompts()
    for _, line := range lines {
        if line == "" {
            continue
        }
        parts := strings.Split(line, ",")
        if len(parts) != len(prompts)+1 {
            continue
        }
        model := parts[0]
        var passes []bool
        for _, passStr := range parts[1:] {
            passes = append(passes, passStr == "true")
        }
        results[model] = passes
    }
    return results
}

// Handle prompt list page
func promptListHandler(w http.ResponseWriter, r *http.Request) {
    prompts := readPrompts()
    t, _ := template.ParseFiles("templates/prompt_list.html")
    t.Execute(w, prompts)
}

// Handle results page
func resultsHandler(w http.ResponseWriter, r *http.Request) {
    prompts := readPrompts()
    results := readResults()
    // Sort models by total score
    models := make([]string, 0)
    for model := range results {
        models = append(models, model)
    }
    // Calculate total scores and sort
    t, _ := template.ParseFiles("templates/results.html")
    t.Execute(w, struct {
        Prompts  []string
        Results  map[string][]bool
        Models   []string
    }{
        Prompts:  prompts,
        Results:  results,
        Models:   models,
    })
}

// Handle AJAX requests to update results
func updateResultHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    model := r.Form.Get("model")
    promptIndexStr := r.Form.Get("promptIndex")
    passStr := r.Form.Get("pass")
    promptIndex, _ := strconv.Atoi(promptIndexStr)
    pass, _ := strconv.ParseBool(passStr)
    // Update the result in results.csv
    // Recalculate total scores and sort
    // Send back the updated scores
    w.Write([]byte("OK"))
}
```

_`templates/prompt_list.html`_

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Prompt List</title>
  </head>
  <body>
    <h1>Prompt List</h1>
    <ul>
      {{range $index, $prompt := .}}
      <li>
        {{$prompt}}
        <a href="/edit_prompt?index={{$index}}">Edit</a>
        <a href="/delete_prompt?index={{$index}}">Delete</a>
      </li>
      {{end}}
    </ul>
    <h2>Add Prompt</h2>
    <form action="/add_prompt" method="post">
      <input type="text" name="prompt" placeholder="Enter new prompt" />
      <input type="submit" value="Add" />
    </form>
    <a href="/results">Go to Results</a>
  </body>
</html>
```

_`templates/results.html`_

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Results</title>
    <script>
      function updateResult(model, promptIndex, pass) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "/update_result", true);
        xhr.setRequestHeader(
          "Content-type",
          "application/x-www-form-urlencoded",
        );
        xhr.onreadystatechange = function () {
          if (xhr.readyState == 4 && xhr.status == 200) {
            // Handle response if needed
          }
        };
        var params =
          "model=" +
          encodeURIComponent(model) +
          "&promptIndex=" +
          promptIndex +
          "&pass=" +
          pass;
        xhr.send(params);
      }
    </script>
  </head>
  <body>
    <h1>Results</h1>
    <table border="1">
      <thead>
        <tr>
          <th>Model</th>
          {{range $index, $prompt := .Prompts}}
          <th>Prompt {{$index + 1}}</th>
          {{end}}
          <th>Total</th>
        </tr>
      </thead>
      <tbody>
        {{range $model, $results := .Results}}
        <tr>
          <td>{{$model}}</td>
          {{range $index, $pass := $results}}
          <td>
            <input
              type="checkbox"
              {{if
              $pass}}checked{{end}}
              onchange="updateResult('{{$model}}', {{$index}}, this.checked)"
            />
          </td>
          {{end}}
          <td>{{len (filter $results true)}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
    <a href="/prompts">Go to Prompt List</a>
  </body>
</html>
```

**Notes:**

- The code above is a simplified version to illustrate the structure.
- Error handling is minimal for brevity.
- The `updateResultHandler` needs to implement the logic to update `results.csv`.
- The sorting of models by total score is not fully implemented.
- The `filter` function in the template is hypothetical; you would need to implement it or calculate the total score in the template.
- Ensure that all files and directories have the correct permissions.
- This is a starting point; further development is needed for a complete application.
