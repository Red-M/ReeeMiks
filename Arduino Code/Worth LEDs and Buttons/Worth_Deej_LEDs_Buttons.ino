const int NUM_SLIDERS = 8;  // Change to correlate with the number of sliders or knobs you intend to use
const int analogInputs[NUM_SLIDERS] = {A3, A2, A1, A0, A4, A5, A6, A7};
const int NUM_BUTTONS = 2; // Change to correlate with the number of buttons you intend to use
const int buttonInputs[NUM_BUTTONS] = {6, 7}; // Change to correlate to the pin numbers listed on your board (i.e D6, D7) and add any as needed

int analogSliderValues[NUM_SLIDERS];
int buttonValues[NUM_BUTTONS];

const int ledPins[2] = {2, 3}; // Array to store LED pins. Change to corrleated to the pin numbers your LEDs are connected to
int ledStates[2] = {LOW, LOW}; // Array to store current LED states
int buttonStates[2] = {LOW, LOW}; // Array to store previous button states
unsigned long debounceTime[2] = {0, 0}; // Debounce timers for buttons
const int debounceDelay = 20; // Debounce delay in milliseconds. Adjust if you feel the LEDs turn on too slowly or too quickly

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(buttonInputs[i], INPUT_PULLUP);
  }

  for (int i = 0; i < 2; i++) {
    pinMode(ledPins[i], OUTPUT); // Set LED pins as output
  }

  Serial.begin(9600);
}

void loop() {
  updateSliderValues();
  updateButtonStates();
  sendSliderValues(); // Send slider & button data
  // printSliderValues(); // For debug
  delay(10);
}

void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
    for (int i = 0; i < NUM_BUTTONS; i++) {
     buttonValues[i] = digitalRead(buttonInputs[i]);
    }
}

void updateButtonStates() {
  for (int i = 0; i < 2; i++) {
    int buttonReading = digitalRead(buttonInputs[i]);
    if (buttonReading == LOW && buttonStates[i] == HIGH && millis() - debounceTime[i] > debounceDelay) {
      // Toggle only the corresponding LED state and update button state
      ledStates[i] = !ledStates[i];
      digitalWrite(ledPins[i], ledStates[i]);
      debounceTime[i] = millis();
      buttonStates[i] = LOW;
    } else if (buttonReading == HIGH) {
      // Button released, reset button state for debounce
      buttonStates[i] = HIGH;
    }
  }
}

void sendSliderValues() {
  String builtString = String("");

  for (int i = 0; i < NUM_SLIDERS; i++) {
    builtString += "s";
    builtString += String((int)analogSliderValues[i]);
    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }

  if (NUM_BUTTONS > 0) {
    builtString += String("|");
  }

  for (int i = 0; i < 2; i++) { // Send data for first 2 buttons only
    builtString += "b";
    builtString += String((int)buttonValues[i]);
    if (i < 2 - 1) {
      builtString += String("|");
    }
  }

  Serial.println(builtString);
}

// Optional debug function (comment out in production)
void printSliderValues() {
  // ... (same as before)
}
