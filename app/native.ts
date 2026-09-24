import { NativeModules } from 'react-native';

export type VpnState = 'off' | 'starting' | 'on';

interface ObxodNative
{
	start(): void;
	stop(): void;
}

const native = NativeModules.Obxod as ObxodNative | undefined;

export function nativeReady(): boolean
{
	return native !== undefined;
}

export function startVpn(): void
{
	native?.start();
}

export function stopVpn(): void
{
	native?.stop();
}
