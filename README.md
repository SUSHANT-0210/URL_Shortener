# Quiz Project

This is a Go program that reads a quiz from a CSV file and presents it to the user. The quiz tracks correct and incorrect answers and supports customizable options like time limits and shuffling questions.

## Features

1. **CSV-Based Quiz**:
   - Reads quiz questions and answers from a CSV file.
   - Default file: `problems.csv`.

2. **Customizable Options**:
   - Set a custom CSV file with the `-csv` flag.
   - Add a time limit for the quiz with the `-limit` flag.

3. **Real-Time Feedback**:
   - Tracks the number of correct and incorrect answers during the quiz.
   - Displays results at the end.

4. **Bonus Features**:
   - Answers are trimmed and case-insensitive to handle formatting variations.
   - Supports shuffling questions with the `-shuffle` flag.

## How to Use

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/SUSHANT-0210/Quiz.git
   cd Quiz

## Build the Program:

go build .

## Run the Program:

./quiz_project.exe -csv=<file_path> -limit=<time_limit> -shuffle=<bool>

-csv: Path to the CSV file. Default is problems.csv.

-limit: Time limit for the quiz in seconds. Default is 30 seconds.

-shuffle: Optional flag to shuffle questions.




