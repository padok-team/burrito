import ThemeProvider from "@/contexts/ThemeContext";

import Layers from "@/pages/layers/Layers";

function App() {
  return (
    <ThemeProvider>
      <Layers />
    </ThemeProvider>
  );
}

export default App;
