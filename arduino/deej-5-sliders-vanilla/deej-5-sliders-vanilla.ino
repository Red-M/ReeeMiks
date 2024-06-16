const int NUM_SLIDERS = 5;
const int sensor_history = 5;
const int slider_range = 2;
const int analogInputs[NUM_SLIDERS] = {A0, A1, A2, A3, A10};
int lastValues[NUM_SLIDERS][sensor_history] = {
  {0,0,0,0,0},
  {0,0,0,0,0},
  {0,0,0,0,0},
  {0,0,0,0,0},
  {0,0,0,0,0}};
String last_string = "";

int analogSliderValues[NUM_SLIDERS];

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  Serial.begin(9600);
}

void loop() {
  updateSliderValues();
  sendSliderValues(); // Actually send data (all the time)
  // printSliderValues(); // For debug
  delay(1);
}

void updateSliderValues() {
  int current_value = 0;
  for (int i = 0; i < NUM_SLIDERS; i++) {
    current_value = analogRead(analogInputs[i]);
    int updateOutput = 1;
    for(int ii = 0 ; ii<sensor_history ; ii++ ){
      // If this historic value doesn't match the current reading,
      // we will not update the output value
      if( !(current_value > lastValues[i][ii] + slider_range || current_value < lastValues[i][ii] - slider_range) ){
        updateOutput = 0;
      }
      // Shift the array elements to make room for new value
      if( ii>0 ){
        lastValues[i][ii-1] = lastValues[i][ii];
      }
    }
    // Update if needed
    if( updateOutput == 1 ){
      lastValues[i][sensor_history-1] = current_value;
      analogSliderValues[i] = current_value;
    }
    // Append the new value
  }
}

void sendSliderValues() {
  String builtString = String("");

  for (int i = 0; i < NUM_SLIDERS; i++) {
    builtString += String((int)analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }

  if (builtString != last_string) {
    Serial.println(builtString);
    last_string = builtString;
  }
}

void printSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    String printedString = String("Slider #") + String(i + 1) + String(": ") + String(analogSliderValues[i]) + String(" mV");
    Serial.write(printedString.c_str());

    if (i < NUM_SLIDERS - 1) {
      Serial.write(" | ");
    } else {
      Serial.write("\n");
    }
  }
}
