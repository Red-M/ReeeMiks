# Summary
This fork includes support for LEDs to indicate when a non-latching pushbutton is pressed, specifically to indicate an audio source being muted, and some additional guidance for using deej. However, I am operating under the assumption you've already gone through the process of choosing hardware and have begun building. If you need additional build support, refer to the [deej FAQ](https://github.com/omriharel/deej/blob/master/docs/faq/faq.md). This is a fork of Miodec's deej fork so it of course includes their addition of remappable button support. Buttons must be defined in config with int values. The key values can be found here: https://github.com/micmonay/keybd_event/blob/master/keybd_windows.go (make sure to convert hex values to int)

Be sure to visit the [original repository](https://github.com/omriharel/deej)

# How To
 - Download the `Worth_Deej_LEDs_Buttons.ino` file from the [release section](https://github.com/wildmanworth/deej/releases/tag/v0.1) and upload to your board using Arduino IDE
 - Download the `deej-release.exe` file and `config.yaml` from the [release section](https://github.com/wildmanworth/deej/releases/tag/v0.1) and place in folder together.
 - Edit both the .ino file and config.yaml file to reflect your board pin configuration, apps controlled, button mapping, and com_port
 - Refer to [Fritzing Wiring Sketch](https://github.com/eric-worth/deej/blob/master/Fritzing%20Wiring%20Sketch.jpg) for guidance in wiring. You can add as many buttons, LEDs, and potentiometers as your Arduino will allow!
 - Launch deej!
 - See deej [FAQ](https://github.com/omriharel/deej/blob/master/docs/faq/faq.md) for additional guidance
 - Join, contribute to, help others, and ask for help in the [Deej Discord Server](https://discord.gg/nf88NJu)

# Case files

Original case files coming soon

# Build

Build log coming soon



