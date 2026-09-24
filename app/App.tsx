import { useState } from 'react';
import { Pressable, StatusBar, StyleSheet, Text, useColorScheme, View } from 'react-native';
import { SafeAreaProvider, SafeAreaView } from 'react-native-safe-area-context';
import { Home } from './screens/home';
import { Rules } from './screens/rules';
import { Log } from './screens/log';

type Screen = 'home' | 'rules' | 'log';

function App()
{
	const scheme = useColorScheme();
	const [screen, setScreen] = useState<Screen>('home');

	return (
		<SafeAreaProvider>
			<StatusBar barStyle={scheme === 'dark' ? 'light-content' : 'dark-content'} />
			<SafeAreaView style={styles.screen}>
				{screen === 'home' ? <Home /> : null}
				{screen === 'rules' ? <Rules /> : null}
				{screen === 'log' ? <Log /> : null}
				<View style={styles.tabs}>
					<Pressable style={styles.tab} onPress={() => setScreen('home')}>
						<Text>home</Text>
					</Pressable>
					<Pressable style={styles.tab} onPress={() => setScreen('rules')}>
						<Text>rules</Text>
					</Pressable>
					<Pressable style={styles.tab} onPress={() => setScreen('log')}>
						<Text>log</Text>
					</Pressable>
				</View>
			</SafeAreaView>
		</SafeAreaProvider>
	);
}

const styles = StyleSheet.create(
{
	screen:
	{
		flex: 1,
		backgroundColor: '#ffffff',
	},
	tabs:
	{
		flexDirection: 'row',
		borderTopWidth: 1,
		borderTopColor: '#000000',
	},
	tab:
	{
		flex: 1,
		alignItems: 'center',
		paddingVertical: 14,
	},
});

export default App;
