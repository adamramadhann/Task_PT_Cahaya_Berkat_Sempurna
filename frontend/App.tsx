import {StatusBar} from 'react-native';
import {SafeAreaProvider} from 'react-native-safe-area-context';

import {TaskListScreen} from './src/screens/TaskListScreen';

function App() {
  return (
    <SafeAreaProvider>
      <StatusBar barStyle="dark-content" />
      <TaskListScreen />
    </SafeAreaProvider>
  );
}

export default App;
