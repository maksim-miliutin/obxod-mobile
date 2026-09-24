import { Pressable, StyleSheet, Text, View } from 'react-native';

const rules = ['discord.com = fake', 'youtube.com = cut', 'x.com = cut,ttl:4', 'telegram.org = disorder'];

export function Rules()
{
	return (
		<View style={styles.body}>
			<Text style={styles.title}>rules</Text>
			{rules.map((line) => (
				<Text key={line} style={styles.row}>{line}</Text>
			))}
			<Pressable style={styles.add}>
				<Text>add</Text>
			</Pressable>
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
	add:
	{
		marginTop: 8,
		alignSelf: 'flex-start',
		borderWidth: 1,
		borderColor: '#000000',
		paddingHorizontal: 16,
		paddingVertical: 10,
	},
});
