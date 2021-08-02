# Go Noise

## System Requirements 

This is work-in-progress alpha software. The only current build is for macOS (I'm running macOS Big Sur). Eventually, I will create builds for multiple platforms, but for now please use this release at your own risk. Thank you!

## Setup

By running the `noise` executable, a Virtual MIDI port will be created called on your system called `NoiseVirtualOut`.

To work with the app, open the DAW of your choice and select `NoiseVirtualOut` as your MIDI device.     

---

**Note**: If you are using Reaper (and possibly other DAWs as well), you have to "remind" Reaper about the app if you opened Reaper first, or if you closed the app and reopened it. 

In Reaper's preferences, in the Audio > MIDI Devices section, click the button `[Reset all MIDI devices]` - this should cause Repear to recognize the app as a device. 

---

## Editing content

There are two primary sections of the app: The `grid` and the `control board`. 

---

### Grid 

The `grid` is a midi note sequencer. 

- Black cells are generated by the system
- Blue cells are created by the user
- Grey cells deactivated by the user (both system and user cells can be deactivated).

To `activate` a cell, `left-click` any empty cell. 

To `deactivate` a cell, `right-click` any active cell.

To `delete` a user-created cell, right-click to deactivate then right-click again to delete.
 
---

### Control board

To be written...
