# Worklog

A plugin for the daiv CLI tool.

## Installation

### From Source

1. Clone the repository:
   ```
   git clone https://github.com/iures/worklog.git
   cd worklog
   ```

2. Build the plugin:
   ```
   go build -buildmode=plugin -o worklog.so
   ```

3. Install the plugin:
   ```
   daiv plugin install /path/to/worklog.so
   ```

### From GitHub

```
daiv plugin install iures/worklog
```

## Configuration

This plugin requires the following configuration:

- worklog.apikey: API key for the service

You can configure these settings when you first run daiv after installing the plugin.

## Usage

After installation, the plugin will be automatically loaded when you start daiv.

## Development

1. Fork this repository
2. Make your changes
3. Build and test locally
4. Submit a pull request
