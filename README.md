# 🏆 LLM Tournament Arena

**A comprehensive benchmarking platform for evaluating and comparing Large Language Models**  
*Real-time scoring • Test suite management • Collaborative evaluation • Advanced analytics*

📦 **Single Binary Deployment** • ⚡ **WebSocket Real-Time Updates** • 📊 **Interactive Dashboards**

<details>
    <summary>Program Screenshots (expand)</summary>

UI Results:
![UI Results](./assets/ui-results.png)
UI Evaluate:
![UI Evaluate](./assets/ui-evaluate.png)
UI Stats:
![UI Stats](./assets/ui-stats.png)
UI Prompts:
![UI Prompts](./assets/ui-prompts.png)
UI Edit Prompt:
![UI Edit Prompt](./assets/ui-edit-prompt.png)
UI Profiles:
![UI Profiles](./assets/ui-profiles.png)

</details>

## 🚀 Quick Start

```bash
# Clone & Run
git clone https://github.com/lavantien/llm-tournament.git
cd llm-tournament
make run
```

Access at `http://localhost:8080`

## 🌟 Key Features

### 🧪 **Evaluation Engine**
- 🎯 Real-time scoring with WebSocket updates (0-100 scale with 5 levels)
- 📈 Automatic model ranking with real-time leaderboard
- 🧮 Granular scoring system (0/5, 1/5, 2/5, 3/5, 4/5, 5/5)
- 📉 Pass percentage calculations and visualization
- 🔄 Instant updates across all connected clients
- 🔀 Random score generation for prototyping
- ⏪ State backup and restore functionality

### 📚 **Test Suite Management**
- 🗂️ Create/rename/delete independent prompt suites
- 🔗 Isolated profiles and results per suite
- ⚡ One-click suite switching with instant UI updates
- 📦 Complete suite export/import (JSON)
- 🏷️ Profile-based prompt categorization and filtering

### ✍️ **Prompt Workshop**
- 📝 Rich Markdown editing with live preview
- 🖇️ Profile assignment for prompt categorization
- 🧩 Bulk selection, deletion, and export operations
- 🎚️ Drag-and-drop reordering with automatic saving
- 🔍 Real-time search and multi-criteria filtering
- 📋 One-click copy functionality for prompt text
- 📤 JSON export/import with validation

### 🤖 **Model Arena**
- ➕ Quick model addition with automatic score initialization
- ✏️ In-place model renaming with result preservation
- 🗑️ Model deletion with confirmation
- 📊 Color-coded scoring visualization (red to blue gradient)
- 🔄 Consistent state persistence across sessions
- 🔍 Model search and filtering capabilities

### 👤 **Profile System**
- 📋 Create reusable evaluation profiles
- 🔖 Associate profiles with prompts for categorization
- 🔄 Automatic prompt updates when profiles are renamed
- 🔍 Profile-based filtering in prompt view
- 📝 Markdown description support with preview

### 📊 **Analytics Suite**
- 📊 Detailed score breakdowns with Chart.js visualizations
- 🏆 Comprehensive tier classification system:
  - Transcendent (1900-2000) 🌌
  - Super-Grandmaster (1800-1899) 🌟
  - Grandmaster (1700-1799) 🥇
  - International Master (1600-1699) 🎖️
  - Master (1500-1599) 🏅
  - Expert (1400-1499) 🎓
  - Pro Player (1200-1399) 🎮
  - Advanced Player (1000-1199) 🎯
  - Intermediate Player (800-999) 📈
  - Veteran (600-799) 👨‍💼
  - Beginner (0-599) 🐣
- 📈 Score distribution visualization
- 📋 Tier-based model grouping
- 📑 Performance comparison across models

### 💻 **Evaluation Interface**
- 🎯 Streamlined scoring with color-coded buttons
- 📝 Full prompt and solution display with Markdown rendering
- ⬅️➡️ Previous/Next navigation between prompts
- 📋 One-click copying of raw prompt text
- 🔍 Clear visualization of current scores
- 🏃‍♂️ Rapid evaluation workflow

### 🔄 **Real-Time Collaboration**
- 🌐 WebSocket-based instant updates across all clients
- 📤 Simultaneous editing with conflict resolution
- 🔄 Broadcast of all changes to connected users
- 📡 Connection status monitoring
- 🔄 Automatic reconnection handling

## 🛠️ Tech Stack

**Backend**  
`Go 1.21+` • `Gorilla WebSocket` • `Blackfriday` • `Bluemonday`

**Frontend**  
`HTML5` • `CSS3` • `JavaScript ES6+` • `Chart.js 4.x` • `Marked.js`

**Data**  
`JSON Storage` • `File-based Persistence` • `JSON Import/Export` • `State Versioning`

**Security**  
`XSS Sanitization` • `CORS Protection` • `Input Validation` • `Error Handling`

## 🧰 Complementary Tools

**Text-to-Speech**  
`tools/tts/podcast.py` - Generate podcast audio from text scripts using Kokoro ONNX models

**Background Removal**  
`tools/bg_batch_eraser/main.py` - Remove backgrounds from images using BEN2 model  
`tools/bg_batch_eraser/vidseg.py` - Extract foreground from videos with alpha channel support  
`tools/bg_batch_eraser/BEN2.py` - Core background eraser neural network implementation

**LLM Integration**  
`tools/openwebui/pipes/anthropic_claude_thinking_96k.py` - OpenWebUI pipe for Claude with thinking mode (96k context)  
`tools/ragweb_agent` - RAG capabilities for web-based content

## 🏁 Getting Started

### Prerequisites
- Go 1.21+
- Make
- Git

### Installation & Running
```bash
# Development mode
./dev.sh

# Production build
make build
./release/llm-tournament
```

## 📚 Usage Guide

1. **Set Up Test Suites**
   - Create a new suite for your evaluation task
   - Configure profiles for different prompt categories
   - Import existing prompts or create new ones

2. **Configure Models**
   - Add each model you want to evaluate
   - Models can represent different LLMs, versions, or configurations

3. **Prepare Prompts**
   - Write prompts with appropriate solutions
   - Assign profiles for categorization
   - Arrange prompts in desired evaluation order

4. **Run Evaluations**
   - Navigate through prompts and assess each model
   - Use the 0-5 scoring system (0, 20, 40, 60, 80, 100 points)
   - Copy prompts directly to your LLM for testing

5. **Analyze Results**
   - View the results page for summary scores
   - Examine tier classifications in the stats page
   - Compare performance across different prompt types
   - Export results for external analysis

## 🔧 Advanced Features

- **Bulk Operations**: Select multiple prompts for deletion or other actions
- **Drag-and-Drop**: Reorder prompts with intuitive drag-and-drop interface
- **State Preservation**: Previous state can be restored with the "Previous" button
- **Mock Data**: Generate random scores to prototype and test visualizations
- **Search & Filter**: Find specific prompts, models, or profiles quickly

## 🤝 Contribution

We welcome contributions!  
📌 First time? Try `good first issue` labeled tickets  
🔧 Core areas needing help:
- Evaluation workflow enhancements
- Additional storage backends
- Advanced visualization
- CI/CD pipeline improvements

**Contribution Process**:
1. Fork repository
2. Create feature branch
3. Submit PR with description
4. Address review comments
5. Merge after approval

## 🗺 Roadmap

### Q2 2025
- 🧠 Multi-LLM consensus scoring
- 🌐 Distributed evaluation mode
- 🔍 Advanced search syntax
- 📱 Responsive mobile design

### Q3 2025
- 📊 Custom metric definitions
- 🤖 Auto-evaluation agents
- 🔄 CI/CD integration
- 🔐 User authentication

## 📜 License

MIT License - See [LICENSE](LICENSE) for details

## 📬 Contact

My work email: [cariyaputta@gmail.com](mailto:cariyaputta@gmail.com)
