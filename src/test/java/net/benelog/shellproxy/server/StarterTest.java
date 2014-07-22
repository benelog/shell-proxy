/*
 * @(#)StarterTest.java $version 2012. 8. 30.
 *
 * Copyright 2007 NHN Corp. All rights Reserved. 
 * NHN PROPRIETARY/CONFIDENTIAL. Use is subject to license terms.
 */

package net.benelog.shellproxy.server;

import static org.hamcrest.CoreMatchers.*;
import static org.junit.Assert.*;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import org.junit.Test;


/**
 * @author benelog
 */
public class StarterTest {
	final Starter starter = new Starter();

	@Test
	public void defaultPortShoulBeUsedWhenArgumentIsEmpty() {
		String[] args = new String[]{};
		
		int port = starter.parsePort(args);
		
		assertThat(port,is(Starter.DEFAULT_PORT));
	}

	
	@Test
	public void defaultPortShoulBeUsedWhenArgumentIsNotNumber() {
		String[] args = new String[]{"abc"};
		
		int port = starter.parsePort(args);
	
		assertThat(port,is(Starter.DEFAULT_PORT));
	}
	
	@Test
	public void portShouldBeParsedWhenArguementIsNumber() {
		String[] args = new String[]{"18083"};
		
		int port = starter.parsePort(args);
	
		assertThat(port,is(18083));
	}
	
	
	@Test
	public void starterShuldBeRunWhenArgmentIsNull() throws Exception {
		Starter.main(null);
	}
	
	@Test
	public void starterShouldBeRunWhenArgmentIsEmpty() throws Exception {
		Starter.main(new String[]{});
	}
	
	@Test
	public void serverShouldBeStartedAndStop() throws Exception {
		
		final AtomicBoolean interupted = new AtomicBoolean(false);
		Runnable startAction = new Runnable(){
			@Override
			public void run() {
				try {
					Starter.main(new String[]{"12300", "echo"});
				} catch (InterruptedException e) {
					interupted.set(true);
					e.printStackTrace();
				} catch (Exception e) {
					e.printStackTrace();
				}	
			}
		};
		ExecutorService exeuctor = Executors.newSingleThreadExecutor();
		exeuctor.execute(startAction);
		TimeUnit.SECONDS.sleep(1);
		exeuctor.shutdownNow();
		TimeUnit.SECONDS.sleep(1);

		assertThat(interupted.get(), is(true));
	}
}
