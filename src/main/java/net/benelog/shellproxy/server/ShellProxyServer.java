package net.benelog.shellproxy.server;

import net.benelog.shellproxy.servlet.ExecutionServlet;
import net.benelog.shellproxy.servlet.StopServlet;

import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;


public class ShellProxyServer {
	
	Server server;
	public ShellProxyServer(int port) {
		server = new Server(port);
		ServletContextHandler context = new ServletContextHandler(	ServletContextHandler.NO_SESSIONS);
		context.setContextPath("/");
		context.addServlet(new ServletHolder(new ExecutionServlet()),"/");
		context.addServlet(new ServletHolder(new StopServlet()),"/stop");
		server.setHandler(context);
	}

	public void start() throws Exception, InterruptedException {
		server.start();
		server.join();
	}
	
	public void stop() throws Exception{
		server.stop();
	}
}
