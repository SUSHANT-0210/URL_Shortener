# Quiz Project

A command-line quiz application built in Go that reads questions from a CSV file and challenges users to answer them within a time limit.

## Features

- **CSV-Based Quiz**: Reads questions and answers from a customizable CSV file
- **Time Limit**: Configurable countdown timer adds an element of challenge
- **Question Shuffling**: Optional randomization of question order
- **Simple Interface**: Clean command-line interface with clear feedback
- **Score Tracking**: Tracks and displays final performance statistics

## Installation

```bash
# Clone the repository
git clone https://github.com/SUSHANT-0210/Quiz.git
cd Quiz

# Build the program
go build -o quiz
```

## Usage

```bash
./quiz -csv=filename.csv -limit=seconds -shuffle=true|false
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-csv` | Path to the CSV file containing questions | `problems.csv` |
| `-limit` | Time limit for the quiz in seconds | `30` |
| `-shuffle` | Whether to randomize question order | `false` |

### CSV File Format

The CSV file should follow this format:
```
question1,answer1
question2,answer2
...
```

Example:
```
5+5,10
capital of france,paris
7*8,56
```

## Example

```bash
# Run with default settings
./quiz

# Run with custom settings
./quiz -csv=math_problems.csv -limit=60 -shuffle=true
```

## Development

Requirements:
- Go 1.16 or higher
