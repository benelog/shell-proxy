package net.benelog.shellproxy.server;

import java.net.InetAddress;
import java.net.UnknownHostException;

/**
 * @author benelog@gmail.com
 */
public class Starter {
	static final int DEFAULT_PORT = 18080;
	public static void main(String[] args) throws Exception {
		Starter starter = new Starter();
		starter.launchServer(args);
	}

	private ShellProxyServer server;

	void launchServer(String[] args)
			throws UnknownHostException, Exception, InterruptedException {
		printUsage();
		if(args != null && args.length > 0 ) {
			int port = parsePort(args);
			printServerAddressInfo(port);
			this.server = new ShellProxyServer(port);
			addShutdownHook(server);
			this.server.start();
		}
	}
	
	int parsePort(String[] args) {
		if (args.length == 0) {
			return DEFAULT_PORT;
		}

		try {
			return Integer.parseInt(args[0]);
		} catch (NumberFormatException nfe) {
			System.out.println("Cannot parse port : " + args[0]);
			return DEFAULT_PORT;
		}
	}

	private void printServerAddressInfo(int port) throws UnknownHostException {
		InetAddress localhost = InetAddress.getLocalHost();
		String ip = localhost.getHostAddress();
		System.out.println("Web address: http://" + ip + ":" + port);
	}

	private void addShutdownHook(final ShellProxyServer server) {
		Runtime.getRuntime().addShutdownHook(new Thread() {
			public void run() {
				try {
					System.out.println("Server stop");
					server.stop();
				} catch (Exception e) {
					e.printStackTrace();
				}
			}
		});
	}

	private static void printUsage() {
		System.out.println("-----------------------------");
		System.out.println("Usage:");
		System.out.println("   Prompt>java -jar shell-proxy.jar [port]");
		System.out.println();
		System.out.println("-----------------------------");
		System.out.println();
	}
}
