import { StyleSheet, Text, View } from 'react-native';

const entries = ['discord.com  fake  1.2 MB', 'youtube.com  cut  340 KB', 'x.com  cut,ttl:4  88 KB', 'example.com  passed'];

export function Log()
{
	return (
		<View style={styles.body}>
			<Text style={styles.title}>log</Text>
			{entries.map((line) => (
				<Text key={line} style={styles.row}>{line}</Text>
			))}
		</View>
	);
}

const styles = StyleSheet.create(
{
	body:
	{
		flex: 1,
		padding: 16,
		gap: 12,
	},
	title:
	{
		fontSize: 22,
		fontWeight: '700',
	},
	row:
	{
		fontSize: 15,
	},
});
