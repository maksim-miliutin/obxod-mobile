import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { nativeReady, startVpn, stopVpn } from '../native';
import type { VpnState } from '../native';

export function Home()
{
	const [state, setState] = useState<VpnState>('off');

	function toggle()
	{
		if (state === 'off')
		{
			startVpn();
			setState('on');
			return;
		}

		stopVpn();
		setState('off');
	}

	return (
		<View style={styles.body}>
			<Text style={styles.title}>obxod</Text>
			<Pressable style={styles.button} onPress={toggle}>
				<Text style={styles.buttonText}>{state}</Text>
			</Pressable>
			{!nativeReady() ? <Text style={styles.note}>native module not loaded</Text> : null}
		</View>
	);
}

const styles = StyleSheet.create(
{
	body:
	{
		flex: 1,
		alignItems: 'center',
		justifyContent: 'center',
		gap: 24,
	},
	title:
	{
		fontSize: 22,
		fontWeight: '700',
	},
	button:
	{
		width: 150,
		height: 150,
		borderRadius: 75,
		borderWidth: 1,
		borderColor: '#000000',
		alignItems: 'center',
		justifyContent: 'center',
	},
	buttonText:
	{
		fontSize: 22,
	},
	note:
	{
		fontSize: 12,
		color: '#666666',
	},
});
