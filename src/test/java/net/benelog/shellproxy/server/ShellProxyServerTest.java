package net.benelog.shellproxy.server;

import static org.junit.Assert.*;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import net.benelog.shellproxy.server.ShellProxyServer;

import org.junit.Test;


public class ShellProxyServerTest {
	
	@Test
	public void startAndStop() throws InterruptedException, Exception {
		// given 
		final ShellProxyServer server = new ShellProxyServer(10023);
		
		Runnable startAction = new Runnable(){
			@Override
			public void run() {
				try {
					server.start();
				} catch (InterruptedException e) {
					fail("interupted");
				} catch (Exception e) {
					fail();
				}			
			}
			
		};
		
		ExecutorService exeuctor = Executors.newSingleThreadExecutor();
		exeuctor.execute(startAction);
		TimeUnit.SECONDS.sleep(1);
		server.stop();
	}

}
