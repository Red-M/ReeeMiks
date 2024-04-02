#Summary
This fork includes some additional guidance for using deej, as well as LED supports to indicate when a button is pressed, specifically to indicate an audio source being muted. This is a fork of Miodec's deej fork so it of course includes their addition of remappable button support. Buttons must be defined in config with int values. The key values can be found here: https://github.com/micmonay/keybd_event/blob/master/keybd_windows.go (make sure to convert hex values to int)

Be sure to visit the [original repository](https://github.com/omriharel/deej)

# How To
 - Download the `Worth_Deej_LEDs_Buttons.ino` file from the [release section](https://github.com/wildmanworth/deej/releases/tag/v0.1) and upload to your board using Arduino IDE
 - Download the `deej-release.exe` file and `config.yaml` from the [release section](https://github.com/wildmanworth/deej/releases/tag/v0.1) and place in folder together.
 - Edit both the .ino file and config.yaml file to reflect your board pin configuration, apps controlled, button mapping, and com_port
 - Launch deej!
 - See deej [FAQ](https://github.com/omriharel/deej/blob/master/docs/faq/faq.md) for additional guidance
 - Join, contribute to, help others, and ask for help in the [Deej Discord Server](https://discord.gg/nf88NJu)

# Case files

Case files available in the [/models](https://github.com/Miodec/deej/tree/master/models) directory

Original case files coming soon

# Build

Build log coming soon



