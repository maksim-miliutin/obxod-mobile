import { StatusBar, StyleSheet, Text, useColorScheme, View } from 'react-native';
import { SafeAreaProvider, SafeAreaView } from 'react-native-safe-area-context';

function App()
{
	const scheme = useColorScheme();

	return (
		<SafeAreaProvider>
			<StatusBar barStyle={scheme === 'dark' ? 'light-content' : 'dark-content'} />
			<SafeAreaView style={styles.screen}>
				<View style={styles.body}>
					<Text style={styles.title}>obxod-mobile</Text>
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
	},
	body:
	{
		flex: 1,
		alignItems: 'center',
		justifyContent: 'center',
	},
	title:
	{
		fontSize: 20,
	},
});

export default App;
